# W2 宠物文本对话与 AI Gateway 全栈冒烟（guide §5/§10、MILESTONES W2）。
#
# 前置：deploy 目录已执行 docker compose up -d --build（server 镜像含最新代码）。
# 用法：powershell -File scripts\smoke-pet-chat.ps1
#
# 覆盖：登录 → 建宠 → 会话幂等 → 20 轮对话无重复 → 重放幂等 → 游标分页不重不漏
#       → 错误路径（2302/2303/1001/1002）→ 用量字段（model/token/err_code）。
#
# 实现说明：
# - 用 .NET HttpClient 而非 Invoke-RestMethod：后者在 Windows PowerShell 5.1 下不按 UTF-8 收发。
# - 验证码从容器 Redis 读取（sms:code:{phone}）后完成登录。
# - 每次运行使用随机号段，便于重复执行。
# - 失败路径（AI 兜底文案）由服务端单测覆盖，本脚本只验证成功链路与错误码。

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

# 随机号段，便于重复执行。
$seed = Get-Random -Minimum 10000000 -Maximum 99999990
$phone = "138" + $seed.ToString("D8")

Write-Output "=== 0. 健康检查 ==="
$healthResp = $script:http.GetStringAsync("http://127.0.0.1:8080/healthz").GetAwaiter().GetResult()
Check-True "/healthz ok" ($healthResp -like '*"code":0*') $healthResp

Write-Output "=== 1. 注册登录 ==="
$token = Login-User -Phone $phone
Check-True "登录成功" ($token.Length -gt 0)

Write-Output "=== 2. 创建宠物 ==="
$pet = Invoke-Api -Method "POST" -Path "/pets" -Body @{ name = "语燕"; avatar_id = 1 } -Token $token
Check "创建宠物" 0 $pet.code
Check-True "宠物名中文完好" ($pet.data.name -eq "语燕") "got=$($pet.data.name)"
$petId = [int64]$pet.data.id

Write-Output "=== 3. 会话创建幂等 ==="
$c1 = Invoke-Api -Method "POST" -Path "/pet-conversations" -Body @{ pet_id = $petId } -Token $token
Check "首次创建会话" 0 $c1.code
$c2 = Invoke-Api -Method "POST" -Path "/pet-conversations" -Body @{ pet_id = $petId } -Token $token
Check "再次创建会话" 0 $c2.code
Check-True "会话 id 一致（不产生重复会话）" ([int64]$c1.data.id -eq [int64]$c2.data.id) "$($c1.data.id) vs $($c2.data.id)"
$convId = [int64]$c1.data.id

Write-Output "=== 4. 连续 20 轮对话 ==="
$userIds = New-Object System.Collections.Generic.List[int64]
$clientIds = New-Object System.Collections.Generic.List[string]
for ($i = 1; $i -le 20; $i++) {
    $clientMsgId = [guid]::NewGuid().ToString()
    $res = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
        pet_id        = $petId
        content       = "第 $i 轮：今天也很累"
        client_msg_id = $clientMsgId
    } -Token $token
    if ($res.code -ne 0) {
        Write-Output "  FAIL  第 $i 轮发送失败 (code=$($res.code) msg=$($res.msg))"
        $script:failures++
        break
    }
    if (-not $res.data.assistant_message -or $res.data.assistant_message.content -eq "") {
        Write-Output "  FAIL  第 $i 轮缺少宠物回复"
        $script:failures++
        break
    }
    $userIds.Add([int64]$res.data.user_message.id)
    $clientIds.Add($clientMsgId)
}
Check-True "20 轮全部成功且都有回复" ($userIds.Count -eq 20) "count=$($userIds.Count)"
Check-True "用户消息 id 互不重复" (($userIds | Select-Object -Unique).Count -eq $userIds.Count)
if ($userIds.Count -eq 20) {
    $last = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
        pet_id        = $petId
        content       = "重放测试"
        client_msg_id = $clientIds[0]
    } -Token $token
    Check "重放同一 client_msg_id" 0 $last.code
    Check-True "重放返回首次用户消息" ([int64]$last.data.user_message.id -eq $userIds[0]) "got=$($last.data.user_message.id)"
}

Write-Output "=== 5. 用量字段（mock provider） ==="
if ($userIds.Count -ge 1) {
    $withUsage = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{
        pet_id  = $petId
        content = "看看用量"
    } -Token $token
    Check-True "usage.model 非空" ($withUsage.data.usage.model -ne "") "got=$($withUsage.data.usage.model)"
    Check-True "usage.input_tokens > 0" ([int]$withUsage.data.usage.input_tokens -gt 0) "got=$($withUsage.data.usage.input_tokens)"
    Check-True "成功时 usage.err_code = 0" ([int]$withUsage.data.usage.err_code -eq 0) "got=$($withUsage.data.usage.err_code)"
    Check-True "streaming = false（一期非流式）" (-not $withUsage.data.streaming)
}

Write-Output "=== 6. 游标分页不重不漏 ==="
$page1 = Invoke-Api -Method "GET" -Path "/pet-messages?conv_id=$convId&cursor=0&limit=20" -Token $token
Check "首页拉取" 0 $page1.code
Check-True "首页 20 条且 has_more" (($page1.data.items.Count -eq 20) -and $page1.data.has_more) "count=$($page1.data.items.Count) hasMore=$($page1.data.has_more)"
$seen = @{}
foreach ($m in $page1.data.items) { $seen[[int64]$m.id] = $true }

$cursor = [int64]$page1.data.next_cursor
$dup = $false
while ($true) {
    $page = Invoke-Api -Method "GET" -Path "/pet-messages?conv_id=$convId&cursor=$cursor&limit=20" -Token $token
    if ($page.code -ne 0) { Write-Output "  FAIL  翻页失败 code=$($page.code)"; $script:failures++; break }
    foreach ($m in $page.data.items) {
        if ($seen.ContainsKey([int64]$m.id)) { $dup = $true }
        $seen[[int64]$m.id] = $true
    }
    if (-not $page.data.has_more) { break }
    $cursor = [int64]$page.data.next_cursor
}
Check-True "分页无重复消息" (-not $dup)
# 20 轮 * 2 + 1 条用量测试 + 1 条重放不新增 = 42 条
Check-True "消息总数 = 42" ($seen.Count -eq 42) "count=$($seen.Count)"

Write-Output "=== 7. 错误路径 ==="
$badContent = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{ pet_id = $petId; content = "   " } -Token $token
Check "空白内容 → 2302" 2302 $badContent.code

$badPet = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{ pet_id = 999999; content = "你好" } -Token $token
Check "宠物不存在 → 2303" 2303 $badPet.code

$noToken = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{ pet_id = $petId; content = "你好" }
Check "未认证 → 1002" 1002 $noToken.code

$noConv = Invoke-Api -Method "GET" -Path "/pet-messages" -Token $token
Check "缺少 conv_id → 1001" 1001 $noConv.code

$otherConv = Invoke-Api -Method "GET" -Path "/pet-messages?conv_id=999999" -Token $token
Check "他人/不存在会话 → 2301" 2301 $otherConv.code

Write-Output ""
if ($script:failures -eq 0) { Write-Output "SMOKE RESULT: ALL PASSED" }
else { Write-Output ("SMOKE RESULT: " + $script:failures + " FAILED"); exit 1 }
