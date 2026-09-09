# W4 成长系统全栈冒烟（guide §4、MILESTONES W4）。
#
# 前置：deploy 目录已执行 docker compose up -d --build（server 镜像含最新代码）。
# 用法：powershell -File scripts\smoke-pet-growth.ps1
#
# 覆盖：多轮互动 → 亲密度/等级提升 → 成长事件可查且可解释 → 越权与参数错误。
# 实现约定与既有冒烟脚本一致：.NET HttpClient 保证 UTF-8、随机号段、UTF-8 with BOM。

$ErrorActionPreference = "Continue"
Add-Type -AssemblyName System.Net.Http

$base = "http://127.0.0.1:8080/api/v1"
$script:failures = 0
$script:http = [System.Net.Http.HttpClient]::new()

function Invoke-Api {
    param(
        [string]$Method,
        [string]$Path,
        [hashtable]$Body,
        [string]$Token
    )
    $req = [System.Net.Http.HttpRequestMessage]::new(
        [System.Net.Http.HttpMethod]::new($Method.ToUpper()), "$base$Path")
    if ($Token) { $req.Headers.Add("Authorization", "Bearer $Token") }
    if ($Body) {
        $json = $Body | ConvertTo-Json -Compress
        $req.Content = [System.Net.Http.StringContent]::new(
            $json, [System.Text.Encoding]::UTF8, "application/json")
    }
    $resp = $script:http.SendAsync($req).GetAwaiter().GetResult()
    $raw = $resp.Content.ReadAsStringAsync().GetAwaiter().GetResult()
    $parsed = $null
    try { $parsed = $raw | ConvertFrom-Json } catch {}
    if ($parsed) {
        return @{
            http = [int]$resp.StatusCode
            code = [int]$parsed.code
            msg  = $parsed.msg
            data = $parsed.data
        }
    }
    return @{ http = [int]$resp.StatusCode; code = -1; msg = $raw; data = $null }
}

function Check {
    param([string]$Name, [int]$Expected, $Actual)
    if ($Expected -eq $Actual) { Write-Output "  PASS  $Name (code=$Actual)" }
    else { Write-Output "  FAIL  $Name expected=$Expected actual=$Actual"; $script:failures++ }
}

function Check-True {
    param([string]$Name, [bool]$Cond, [string]$Detail = "")
    if ($Cond) { Write-Output "  PASS  $Name $Detail" }
    else { Write-Output "  FAIL  $Name $Detail"; $script:failures++ }
}

function Login-User {
    param([string]$Phone)
    $null = Invoke-Api -Method "POST" -Path "/auth/sms-code" -Body @{ phone = $Phone }
    $code = (& docker exec yuyan-redis-1 redis-cli get "sms:code:$Phone" | Out-String).Trim()
    $login = Invoke-Api -Method "POST" -Path "/auth/login" -Body @{ phone = $Phone; code = $code }
    if ($login.code -ne 0) { throw "login failed for $Phone : $($login.msg)" }
    return $login.data.access_token
}

$seed = Get-Random -Minimum 10000000 -Maximum 99999980
$phone = "138" + $seed.ToString("D8")
$otherPhone = "138" + ($seed + 1).ToString("D8")

Write-Output "=== 0. 健康检查 ==="
$healthResp = $script:http.GetStringAsync("http://127.0.0.1:8080/healthz").GetAwaiter().GetResult()
Check-True "/healthz ok" ($healthResp -like '*"code":0*') $healthResp

Write-Output "=== 1. 登录与建宠 ==="
$token = Login-User -Phone $phone
$otherToken = Login-User -Phone $otherPhone
$pet = Invoke-Api -Method "POST" -Path "/pets" -Body @{ name = "语燕"; avatar_id = 1 } -Token $token
Check "创建宠物" 0 $pet.code
$petId = [int64]$pet.data.id

Write-Output "=== 2. 初始状态 ==="
$init = Invoke-Api -Method "GET" -Path "/pets/$petId/state" -Token $token
Check-True "初始亲密度为 0" ([int]$init.data.intimacy -eq 0) "got=$($init.data.intimacy)"
Check-True "初始等级为 1" ([int]$init.data.level -eq 1) "got=$($init.data.level)"

Write-Output "=== 3. 多轮互动触发成长 ==="
for ($i = 1; $i -le 10; $i++) {
    $res = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
        pet_id  = $petId
        content = "第 $i 次陪你聊天"
    } -Token $token
    if ($res.code -ne 0) {
        Write-Output "  FAIL  第 $i 轮对话失败 (code=$($res.code) msg=$($res.msg))"
        $script:failures++
        break
    }
}
$state = Invoke-Api -Method "GET" -Path "/pets/$petId/state" -Token $token
Check-True "亲密度随互动增长" ([int]$state.data.intimacy -ge 20) "got=$($state.data.intimacy)"
Check-True "等级提升到 2" ([int]$state.data.level -ge 2) "got=$($state.data.level)"

Write-Output "=== 4. 成长事件可查且可解释 ==="
$events = Invoke-Api -Method "GET" -Path "/pets/$petId/growth-events" -Token $token
Check "查询成长事件" 0 $events.code
Check-True "事件数量 >= 10" ($events.data.Count -ge 10) "count=$($events.data.Count)"
# 同一轮里 level_up 记录在 message 之后（id 更大），最新一条可能是两者之一。
Check-True "最新事件为互动或升级" (($events.data[0].event_type -eq 'message') -or ($events.data[0].event_type -eq 'level_up')) "got=$($events.data[0].event_type)"
Check-True "事件含可解释原因" ($events.data[0].reason -ne '') "got=$($events.data[0].reason)"

$types = $events.data | ForEach-Object { $_.event_type } | Select-Object -Unique
Check-True "包含升级事件" ($types -contains 'level_up') "types=$($types -join ',')"

Write-Output "=== 5. 错误路径 ==="
$noToken = Invoke-Api -Method "GET" -Path "/pets/$petId/growth-events"
Check "未认证 → 1002" 1002 $noToken.code

$badId = Invoke-Api -Method "GET" -Path "/pets/0/growth-events" -Token $token
Check "非法 pet id → 1001" 1001 $badId.code

$other = Invoke-Api -Method "GET" -Path "/pets/$petId/growth-events" -Token $otherToken
Check "他人查看 → 2303" 2303 $other.code

Write-Output ""
if ($script:failures -eq 0) { Write-Output "SMOKE RESULT: ALL PASSED" }
else { Write-Output ("SMOKE RESULT: " + $script:failures + " FAILED"); exit 1 }
