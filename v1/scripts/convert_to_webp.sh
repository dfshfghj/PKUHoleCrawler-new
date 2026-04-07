#!/bin/bash
# 使用 cwebp / gif2webp 将 data/images 下的图片批量转换为 WebP
# 依赖: libwebp (cwebp, gif2webp)
# 用法: ./scripts/convert_to_webp.sh [图片目录] [质量 0-100]

DIR="${1:-data/images}"
QUALITY="${2:-80}"

if [ ! -d "$DIR" ]; then
    echo "错误: 目录 '$DIR' 不存在"
    exit 1
fi

if ! command -v cwebp &>/dev/null; then
    echo "错误: 未找到 cwebp，请先安装 libwebp"
    echo "  Ubuntu/Debian: sudo apt install libwebp-tools"
    echo "  macOS:         brew install webp"
    exit 1
fi

converted=0
skipped=0
failed=0
total_saved=0
total_orig=0

for file in "$DIR"/*; do
    [ -f "$file" ] || continue

    filename=$(basename "$file")
    ext="${filename##*.}"
    name="${filename%.*}"

    # 跳过已经是 webp 的文件
    if [ "$ext" = "webp" ]; then
        ((skipped++))
        continue
    fi

    # 只处理 jpg, png, gif
    case "$ext" in
        jpg|jpeg|png|gif) ;;
        *) ((skipped++)); continue ;;
    esac

    out="$DIR/${name}.webp"

    # 如果目标已存在则跳过
    if [ -f "$out" ]; then
        echo "[跳过] $filename (已存在 ${name}.webp)"
        ((skipped++))
        continue
    fi

    orig_size=$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file" 2>/dev/null)

    if [ "$ext" = "gif" ]; then
        # GIF 使用 gif2webp -mixed 模式（有损+无损混合，兼顾体积和质量）
        if gif2webp -mixed -q "$QUALITY" "$file" -o "$out" 2>/dev/null; then
            new_size=$(stat -c%s "$out" 2>/dev/null || stat -f%z "$out" 2>/dev/null)
            rm -f "$file"
            saved=$(( orig_size - new_size ))
            total_orig=$((total_orig + orig_size))
            total_saved=$((total_saved + saved))
            pct=$(( saved * 100 / orig_size ))
            echo "[GIF→WebP] $filename → ${name}.webp ($orig_size → $new_size bytes, 节省 ${pct}%)"
            ((converted++))
        else
            echo "[失败] $filename (gif2webp 转换失败)"
            rm -f "$out"
            ((failed++))
        fi
    else
        # JPG/PNG 使用 cwebp
        if cwebp -q "$QUALITY" "$file" -o "$out" 2>/dev/null; then
            new_size=$(stat -c%s "$out" 2>/dev/null || stat -f%z "$out" 2>/dev/null)
            rm -f "$file"
            saved=$(( orig_size - new_size ))
            total_orig=$((total_orig + orig_size))
            total_saved=$((total_saved + saved))
            pct=$(( saved * 100 / orig_size ))
            echo "[${ext^^}→WebP] $filename → ${name}.webp ($orig_size → $new_size bytes, 节省 ${pct}%)"
            ((converted++))
        else
            echo "[失败] $filename (cwebp 转换失败)"
            rm -f "$out"
            ((failed++))
        fi
    fi
done

echo ""
echo "========================================="
echo "完成! 转换: $converted, 跳过: $skipped, 失败: $failed"
if [ $total_orig -gt 0 ]; then
    total_pct=$(( total_saved * 100 / total_orig ))
    echo "总计: $((total_orig / 1024)) KB → $(( (total_orig - total_saved) / 1024 )) KB (节省 ${total_pct}%)"
fi
echo "========================================="
