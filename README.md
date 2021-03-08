# i3_swap_biggest
Using i3ipc to swap current window with biggest one on current workspace.

### instal
```bash
go instal swap_biggest.go
```
### usage:
```
# swap with widest ( -b to swap back if trigger one more time)
bindsym $mod+g exec --no-startup-id "swap_biggest -b"
```
