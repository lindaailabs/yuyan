# W3 好友关系全栈 API 冒烟（guide §7 / openspec add-friend-contacts tasks 4.1）。
#
# 前置：deploy 目录已执行 docker compose up -d --build（server 镜像含最新代码）。
# 用法：powershell -File scripts\smoke-contacts.ps1
#
# 覆盖：A→B 申请 → B 列表可见 → 同意 → 双向好友 → 防重复矩阵（2101~2106）
#       → 参数/认证错误（1001/1002）→ 拒绝 → 拒绝后重新申请 → 中文昵称 utf8mb4 往返。
#
# 实现说明：
# - 用 .NET HttpClient 而非 Invoke-RestMethod：后者在 Windows PowerShell 5.1 下不按
#   UTF-8 收发，中文昵称会被写成 '?'。
# - 使用手机号+密码注册/登录，并写入昵称验证 utf8mb4 往返。
# - 每次运行使用随机号段，便于重复执行。
# - 本文件须以 UTF-8 with BOM 保存：PS5.1 按 GBK 读取无 BOM 的 UTF-8，中文会吞掉紧邻引号。

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
    param([string]$Phone, [string]$Nickname)
    $password = "secret123"
    $register = Invoke-Api -Method "POST" -Path "/auth/register" -Body @{ phone = $Phone; password = $password }
    if ($register.code -eq 0) {
        $token = $register.data.access_token
    } elseif ($register.code -eq 2004) {
        $login = Invoke-Api -Method "POST" -Path "/auth/login" -Body @{ phone = $Phone; password = $password }
        if ($login.code -ne 0) { throw "login failed for $Phone : $($login.msg)" }
        $token = $login.data.access_token
    } else {
        throw "register failed for $Phone : $($register.msg)"
    }
    $profile = Invoke-Api -Method "PUT" -Path "/users/me" -Body @{ nickname = $Nickname; avatar_id = 1 } -Token $token
    if ($profile.code -ne 0) { throw "profile failed for $Phone : $($profile.msg)" }
    return @{ token = $token; id = [int64]$profile.data.id }
}

# 每次运行取随机号段，避免与历史数据/限频冲突。
$seed = Get-Random -Minimum 10000000 -Maximum 99999990
$phoneA = "138" + $seed.ToString("D8")
$phoneB = "138" + ($seed + 1).ToString("D8")
$phoneC = "138" + ($seed + 2).ToString("D8")

Write-Output "=== 0. 健康检查 ==="
$healthResp = $script:http.GetStringAsync("http://127.0.0.1:8080/healthz").GetAwaiter().GetResult()
Check-True "/healthz ok" ($healthResp -like '*"code":0*') $healthResp

Write-Output "=== 1. A/B/C 注册登录（中文昵称 utf8mb4 往返） ==="
$a = Login-User -Phone $phoneA -Nickname "语燕A"
$b = Login-User -Phone $phoneB -Nickname "语燕B"
$c = Login-User -Phone $phoneC -Nickname "语燕C"
Write-Output "  A=$($a.id) B=$($b.id) C=$($c.id)"

Write-Output "=== 2. A 搜索 B（脱敏） ==="
$search = Invoke-Api -Method "GET" -Path "/users/search?q=$phoneB" -Token $a.token
Check "搜索 B" 0 $search.code
$masked = $phoneB.Substring(0, 3) + "****" + $phoneB.Substring(7)
Check-True "B 手机号脱敏 $masked" ($search.data[0].phone -eq $masked) "got=$($search.data[0].phone)"
Check-True "B 昵称中文完好" ($search.data[0].nickname -eq "语燕B") "got=$($search.data[0].nickname)"

Write-Output "=== 3. 发起申请 + 防重复矩阵 ==="
$r1 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $b.id } -Token $a.token
Check "A→B 首次申请" 0 $r1.code
$reqId = [int64]$r1.data.id

$r2 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $b.id } -Token $a.token
Check "重复申请→2103" 2103 $r2.code

$r3 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $a.id } -Token $a.token
Check "加自己→2102" 2102 $r3.code

$r4 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = 99999999 } -Token $a.token
Check "目标不存在→2101" 2101 $r4.code

$r5 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = 0 } -Token $a.token
Check "user_id=0→1001" 1001 $r5.code

$r6 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $b.id }
Check "未认证→1002" 1002 $r6.code

Write-Output "=== 4. B 收到申请 ==="
$listB = Invoke-Api -Method "GET" -Path "/friends/requests" -Token $b.token
Check "B 拉取申请列表" 0 $listB.code
Check-True "B 看到 1 条申请" ($listB.data.Count -eq 1) "count=$($listB.data.Count)"
Check-True "申请 id 一致" ([int64]$listB.data[0].id -eq $reqId)
Check-True "申请人昵称=语燕A" ($listB.data[0].from_user.nickname -eq "语燕A") "got=$($listB.data[0].from_user.nickname)"

Write-Output "=== 5. B 同意 → 双向好友 ==="
$acc = Invoke-Api -Method "POST" -Path "/friends/requests/$reqId/accept" -Token $b.token
Check "B 同意" 0 $acc.code
Check-True "状态=accepted(2)" ([int]$acc.data.status -eq 2) "got=$($acc.data.status)"

$fa = Invoke-Api -Method "GET" -Path "/friends" -Token $a.token
$fb = Invoke-Api -Method "GET" -Path "/friends" -Token $b.token
Check-True "A 好友列表含 B" ([int64]$fa.data[0].user.id -eq $b.id)
Check-True "B 好友列表含 A" ([int64]$fb.data[0].user.id -eq $a.id)
Check-True "好友昵称中文完好" ($fa.data[0].user.nickname -eq "语燕B") "got=$($fa.data[0].user.nickname)"

$acc2 = Invoke-Api -Method "POST" -Path "/friends/requests/$reqId/accept" -Token $b.token
Check "二次同意→2106" 2106 $acc2.code

$acc3 = Invoke-Api -Method "POST" -Path "/friends/requests/$reqId/accept" -Token $c.token
Check "C 处理他人申请→2106" 2106 $acc3.code

$r7 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $b.id } -Token $a.token
Check "已是好友再申请→2104" 2104 $r7.code

$listB2 = Invoke-Api -Method "GET" -Path "/friends/requests" -Token $b.token
Check-True "已处理申请不再出现" ($listB2.data.Count -eq 0) "count=$($listB2.data.Count)"

Write-Output "=== 6. C→A 申请 → A 拒绝 → 反向待处理 2105 ==="
$rc = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $a.id } -Token $c.token
Check "C→A 申请" 0 $rc.code
$rcId = [int64]$rc.data.id

$r8 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $c.id } -Token $a.token
Check "A→C 反向待处理→2105" 2105 $r8.code

$rej = Invoke-Api -Method "POST" -Path "/friends/requests/$rcId/reject" -Token $a.token
Check "A 拒绝 C" 0 $rej.code
Check-True "状态=rejected(4)" ([int]$rej.data.status -eq 4) "got=$($rej.data.status)"

$listA = Invoke-Api -Method "GET" -Path "/friends/requests" -Token $a.token
Check-True "A 申请列表已清空" ($listA.data.Count -eq 0) "count=$($listA.data.Count)"

$fa2 = Invoke-Api -Method "GET" -Path "/friends" -Token $a.token
Check-True "A 好友仍只有 B（拒绝不产生好友）" ($fa2.data.Count -eq 1) "count=$($fa2.data.Count)"

$rej2 = Invoke-Api -Method "POST" -Path "/friends/requests/$rcId/reject" -Token $a.token
Check "二次拒绝→2106" 2106 $rej2.code

Write-Output "=== 7. 拒绝后可重新申请 ==="
$rc2 = Invoke-Api -Method "POST" -Path "/friends/requests" -Body @{ user_id = $a.id } -Token $c.token
Check "C 重新申请（rejected→pending）" 0 $rc2.code
$listA2 = Invoke-Api -Method "GET" -Path "/friends/requests" -Token $a.token
Check-True "A 再次看到 C 的申请" ($listA2.data.Count -eq 1) "count=$($listA2.data.Count)"

Write-Output ""
if ($script:failures -eq 0) { Write-Output "SMOKE RESULT: ALL PASSED" }
else { Write-Output ("SMOKE RESULT: " + $script:failures + " FAILED"); exit 1 }
