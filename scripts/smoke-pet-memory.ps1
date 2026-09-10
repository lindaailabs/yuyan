# W3 记忆系统全栈冒烟（guide §4/§5、MILESTONES W3）。
#
# 前置：deploy 目录已执行 docker compose up -d --build（server 镜像含最新代码）。
# 用法：powershell -File scripts\smoke-pet-memory.ps1
#
# 覆盖：说偏好 → 形成记忆 → 后续对话召回并体现 → 删除后列表不可见且不再召回 → 错误路径。
# 实现约定与 smoke-pet-chat.ps1 一致：.NET HttpClient 保证 UTF-8、随机号段、UTF-8 with BOM。

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
        $json = $Body | ConvertTo-Json -Compress -Depth 8
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
    $password = "secret123"
    $register = Invoke-Api -Method "POST" -Path "/auth/register" -Body @{ phone = $Phone; password = $password }
    if ($register.code -eq 0) { return $register.data.access_token }
    if ($register.code -eq 2004) {
        $login = Invoke-Api -Method "POST" -Path "/auth/login" -Body @{ phone = $Phone; password = $password }
        if ($login.code -ne 0) { throw "login failed for $Phone : $($login.msg)" }
        return $login.data.access_token
    }
    throw "register failed for $Phone : $($register.msg)"
}

$seed = Get-Random -Minimum 10000000 -Maximum 99999990
$phone = "138" + $seed.ToString("D8")

Write-Output "=== 0. 健康检查 ==="
$healthResp = $script:http.GetStringAsync("http://127.0.0.1:8080/healthz").GetAwaiter().GetResult()
Check-True "/healthz ok" ($healthResp -like '*"code":0*') $healthResp

Write-Output "=== 1. 注册登录与建宠 ==="
$token = Login-User -Phone $phone
$pet = Invoke-Api -Method "POST" -Path "/pets" -Body @{ name = "语燕"; avatar_id = 1 } -Token $token
Check "创建宠物" 0 $pet.code
$petId = [int64]$pet.data.id

Write-Output "=== 2. 说出偏好 → 形成记忆 ==="
$first = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
    pet_id  = $petId
    content = "我喜欢蓝色"
} -Token $token
Check "发送偏好消息" 0 $first.code
Check-True "本轮形成新记忆" ($first.data.new_memories.Count -ge 1) "count=$($first.data.new_memories.Count)"
$joined = ($first.data.new_memories | ForEach-Object { $_.content }) -join '、'
Check-True "记忆内容包含蓝色" ($joined -like '*蓝色*') "got=$joined"

$list = Invoke-Api -Method "GET" -Path "/pets/$petId/memories" -Token $token
Check "查看记忆列表" 0 $list.code
Check-True "列表含 1 条记忆" ($list.data.Count -eq 1) "count=$($list.data.Count)"
$memoryId = [int64]$list.data[0].id
Check-True "记忆有来源与置信度" (($null -ne $list.data[0].source_msg_id) -and ([double]$list.data[0].confidence -gt 0)) "conf=$($list.data[0].confidence)"

Write-Output "=== 3. 后续对话召回并体现 ==="
$recall = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
    pet_id  = $petId
    content = "你记得我喜欢什么颜色吗"
} -Token $token
Check "再次发送" 0 $recall.code
$reply = $recall.data.assistant_message.content
Check-True "宠物回复体现记忆" ($reply -like '*蓝色*') "got=$reply"

Write-Output "=== 4. 删除记忆后不再召回 ==="
$del = Invoke-Api -Method "DELETE" -Path "/pet-memories/$memoryId" -Token $token
Check "删除记忆" 0 $del.code

$after = Invoke-Api -Method "GET" -Path "/pets/$petId/memories" -Token $token
Check-True "列表不再展示" ($after.data.Count -eq 0) "count=$($after.data.Count)"

$again = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
    pet_id  = $petId
    content = "我喜欢什么颜色呢"
} -Token $token
Check "删除后再次发送" 0 $again.code
$reply2 = $again.data.assistant_message.content
Check-True "回复不再包含已删记忆" ($reply2 -notlike '*蓝色*') "got=$reply2"

Write-Output "=== 5. 错误路径 ==="
$delAgain = Invoke-Api -Method "DELETE" -Path "/pet-memories/$memoryId" -Token $token
Check "重复删除 → 2401" 2401 $delAgain.code

$noToken = Invoke-Api -Method "GET" -Path "/pets/$petId/memories"
Check "未认证 → 1002" 1002 $noToken.code

$badId = Invoke-Api -Method "DELETE" -Path "/pet-memories/0" -Token $token
Check "非法 id → 1001" 1001 $badId.code

Write-Output ""
if ($script:failures -eq 0) { Write-Output "SMOKE RESULT: ALL PASSED" }
else { Write-Output ("SMOKE RESULT: " + $script:failures + " FAILED"); exit 1 }
