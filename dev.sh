#!/bin/bash

SESSION="TestGrid"
APP_DIR="app"

# 自动查找 app 下所有 Go 服务
SERVICES=()

while IFS= read -r dir; do
    SERVICES+=("${dir#"$APP_DIR"/}")
done < <(
    find "$APP_DIR" \
        -name go.mod \
        -type f \
        -not -path "*/vendor/*" \
        -printf '%h\n' |
        sort
)

case "$1" in
  start)
    # 检查 session 是否已经存在
    if tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "tmux session '$SESSION' already exists."
        echo "Use '$0 stop' first, or '$0 restart'."
        exit 1
    fi

    # 没找到服务
    if [ ${#SERVICES[@]} -eq 0 ]; then
        echo "No Go services found in '$APP_DIR'."
        exit 1
    fi

    first=true

    for service in "${SERVICES[@]}"; do
        # iam/api -> iam-api
        window_name="${service//\//-}"

        echo "Starting $service..."

        if [ "$first" = true ]; then
            # 创建 tmux session + 第一个 window
            tmux new-session \
                -d \
                -s "$SESSION" \
                -n "$window_name"

            first=false
        else
            # 后续服务创建 window
            tmux new-window \
                -t "$SESSION" \
                -n "$window_name"
        fi

        # 启动服务
        tmux send-keys \
            -t "$SESSION:$window_name" \
            "cd $APP_DIR/$service && air" \
            C-m
    done

    echo ""
    echo "Started ${#SERVICES[@]} services."
    echo "Attaching to tmux session '$SESSION'..."

    tmux attach-session -t "$SESSION"
    ;;

  stop)
    if tmux has-session -t "$SESSION" 2>/dev/null; then
        tmux kill-session -t "$SESSION"
        echo "Stopped '$SESSION'."
    else
        echo "Session '$SESSION' is not running."
    fi
    ;;

  restart)
    echo "Restarting '$SESSION'..."

    tmux kill-session -t "$SESSION" 2>/dev/null

    "$0" start
    ;;

  *)
    echo "Usage: $0 {start|stop|restart}"
    exit 1
    ;;
esac