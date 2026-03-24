Output demo.gif
Require ffmpeg
Require ttyd

Set Theme "ayu"
Set FontSize 18
Set Width 1200
Set Height 720
Set TypingSpeed 80ms
Set Margin 20
Set Shell "powershell"

Type "cd D:\\Github_Workstation\\GitProjects\\Lenv"
Enter
Sleep 500ms

Type ".\\lenv.exe --help"
Enter
Sleep 2s

Type ".\\lenv.exe release-notes --version v1.0.0"
Enter
Sleep 2s

Type ".\\lenv.exe profile list"
Enter
Sleep 2s

Type ".\\lenv.exe runtime status"
Enter
Sleep 2s

Type ".\\lenv.exe provenance"
Enter
Sleep 2s
