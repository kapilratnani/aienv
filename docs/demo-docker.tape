Output docs/demo-docker.gif
Require aienv

Set Shell bash
Set FontSize 18
Set Width 1200
Set Height 600
Set Theme "Catppuccin Frappe"
Set TypingSpeed 80ms
Set CursorBlink false
Set WindowBar Colorful
Set Padding 40

Env TERM xterm-256color
Env COLORTERM truecolor

# 1. Check Docker is available
Type "aienv docker check"
Sleep 200ms
Enter
Sleep 2s

# 2. Show environment details
Type "aienv show backend-api"
Sleep 200ms
Enter
Sleep 3s

# 3. Activate in Docker sandbox — launches container running opencode
Type "aienv activate --docker backend-api"
Sleep 200ms
Enter
Sleep 15s

# 4. Show loaded skills inside opencode
Type "/skills"
Sleep 300ms
Enter
Sleep 5s

# 5. Show configured MCPs inside opencode
Type "/mcps"
Sleep 300ms
Enter
Sleep 5s

# 6. Exit opencode (container exits too)
Type "/exit"
Sleep 300ms
Enter
Sleep 2s
