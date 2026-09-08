# 生成 8 个预置头像占位图（纯色圆 + 白色数字，128x128 PNG）。
# 用法：powershell -File scripts/gen_avatars.ps1
# 正式插画就绪后直接替换 assets/avatars/avatar_N.png 即可（id 不变）。

Add-Type -AssemblyName System.Drawing

$colors = @('#42A5F5', '#26A69A', '#66BB6A', '#FFA726', '#EF5350', '#AB47BC', '#EC407A', '#5C6BC0')
$outDir = Join-Path $PSScriptRoot "..\apps\app\assets\avatars"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

for ($i = 1; $i -le 8; $i++) {
    $bmp = New-Object System.Drawing.Bitmap 128, 128
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $g.Clear([System.Drawing.Color]::Transparent)

    $brush = New-Object System.Drawing.SolidBrush (
        [System.Drawing.ColorTranslator]::FromHtml($colors[$i - 1]))
    $g.FillEllipse($brush, 4, 4, 120, 120)

    $font = New-Object System.Drawing.Font 'Segoe UI', 56, ([System.Drawing.FontStyle]::Bold)
    $sf = New-Object System.Drawing.StringFormat
    $sf.Alignment = [System.Drawing.StringAlignment]::Center
    $sf.LineAlignment = [System.Drawing.StringAlignment]::Center
    $rect = New-Object System.Drawing.RectangleF 4, 4, 120, 120
    $g.DrawString("$i", $font, [System.Drawing.Brushes]::White, $rect, $sf)

    $g.Dispose()
    $path = Join-Path $outDir "avatar_$i.png"
    $bmp.Save($path, [System.Drawing.Imaging.ImageFormat]::Png)
    $bmp.Dispose()
    Write-Host "generated $path"
}
