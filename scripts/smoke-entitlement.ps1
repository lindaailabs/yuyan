# W5 权益与数据看板全栈冒烟（guide §7、§8、MILESTONES W5）。
#
# 前置：deploy 目录已执行 docker compose up -d --build（server 镜像含最新代码）。
# 用法：powershell -File scripts\smoke-entitlement.ps1
#
# 覆盖：默认免费权益 → 沙盒开通 Pro → 对话消耗额度 → 埋点上报与内部可查 → 越权与参数错误。
# 实现约定与既有冒烟脚本一致：.NET HttpClient 保证 UTF-8、随机号段。

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

$seed = Get-Random -Minimum 10000000 -Maximum 99999980
$phone = "138" + $seed.ToString("D8")

Write-Output "=== 0. 健康检查 ==="
$healthResp = $script:http.GetStringAsync("http://127.0.0.1:8080/healthz").GetAwaiter().GetResult()
Check-True "/healthz ok" ($healthResp -like '*"code":0*') $healthResp

Write-Output "=== 1. 登录 ==="
$token = Login-User -Phone $phone

Write-Output "=== 2. 默认免费权益 ==="
$me = Invoke-Api -Method "GET" -Path "/entitlements/me" -Token $token
Check "查询我的权益" 0 $me.code
Check-True "默认免费版" ($me.data.plan -eq 'free') "plan=$($me.data.plan)"
Check-True "免费版每日 50 条" ([int]$me.data.quota.daily_messages -eq 50) "got=$($me.data.quota.daily_messages)"
Check-True "初始剩余 50" ([int]$me.data.quota.daily_remain -eq 50) "got=$($me.data.quota.daily_remain)"

Write-Output "=== 3. 沙盒开通 Pro ==="
$sb = Invoke-Api -Method "POST" -Path "/entitlements/sandbox-purchase" -Body @{ plan = "pro" } -Token $token
Check "沙盒开通" 0 $sb.code
Check-True "已升级 Pro" ($sb.data.plan -eq 'pro') "plan=$($sb.data.plan)"
Check-True "Pro 每日 500 条" ([int]$sb.data.quota.daily_messages -eq 500) "got=$($sb.data.quota.daily_messages)"

Write-Output "=== 4. 对话消耗额度 ==="
$pet = Invoke-Api -Method "POST" -Path "/pets" -Body @{ name = "语燕"; avatar_id = 1 } -Token $token
Check "创建宠物" 0 $pet.code
$petId = [int64]$pet.data.id
$chat = Invoke-Api -Method "POST" -Path "/pet-messages" -Body @{ pet_id = $petId; content = "今天心情不错" } -Token $token
Check "发送对话" 0 $chat.code
$meAfter = Invoke-Api -Method "GET" -Path "/entitlements/me" -Token $token
Check-True "消耗后 used >= 1" ([int]$meAfter.data.quota.daily_used -ge 1) "used=$($meAfter.data.quota.daily_used)"
Check-True "Pro 剩余 = 500 - used" ([int]$meAfter.data.quota.daily_remain -eq (500 - [int]$meAfter.data.quota.daily_used)) "remain=$($meAfter.data.quota.daily_remain)"

Write-Output "=== 5. 埋点上报与内部可查（非生产）==="
$ev = Invoke-Api -Method "POST" -Path "/events" -Body @{
    events = @(@{ name = "subscription_view"; props = @{ level = "2" } })
} -Token $token
Check "上报埋点" 0 $ev.code
Check-True "接受 1 条" ([int]$ev.data.accepted -eq 1) "got=$($ev.data.accepted)"
$admin = Invoke-Api -Method "GET" -Path "/admin/events?name=subscription_view" -Token $token
Check-True "内部可查埋点" ($admin.http -eq 200) "http=$($admin.http)"

Write-Output "=== 6. 支付回调发放（幂等）==="
$cb = Invoke-Api -Method "POST" -Path "/entitlements/payments/callback" -Body @{ order_no = "smoke-cb-1"; plan = "pro" } -Token $token
Check "支付回调" 0 $cb.code
$cbDup = Invoke-Api -Method "POST" -Path "/entitlements/payments/callback" -Body @{ order_no = "smoke-cb-1"; plan = "pro" } -Token $token
Check "重复回调仍成功" 0 $cbDup.code

Write-Output "=== 7. 错误路径 ==="
$noToken = Invoke-Api -Method "GET" -Path "/entitlements/me"
Check "未认证 → 1002" 1002 $noToken.code

$badPlan = Invoke-Api -Method "POST" -Path "/entitlements/sandbox-purchase" -Body @{ plan = "unknown" } -Token $token
Check "非法套餐 → 1001" 1001 $badPlan.code

$badCb = Invoke-Api -Method "POST" -Path "/entitlements/payments/callback" -Body @{ plan = "pro" } -Token $token
Check "回调缺 order_no → 1001" 1001 $badCb.code

Write-Output ""
if ($script:failures -eq 0) { Write-Output "SMOKE RESULT: ALL PASSED" }
else { Write-Output ("SMOKE RESULT: " + $script:failures + " FAILED"); exit 1 }
