# Kiosk Installation #

## Basic Installation ## 
Install Raspi OS minimal with user ```kiosk```


## Update and install needed Packets ##

**Run updates and install xserver minimal packets**
```
sudo apt-get update
sudo apt-get upgrade
sudo apt-get install xserver-xorg xinit xserver-xorg-video-fbdev lxde lxde-common chromium-browser --yes
```


**Set Boot to Desktop with user logged user 'kiosk'**
- start ```sudo raspi-config```
- select  1 "System Options"
- select S5 "Boot / Auto Login"
- choose B4 "Desktop GUI, automatically logged in as 'kiosk' user"


**Install & Setup Kiosk**
```
wget https://github.com/SebastianRau/kiosk/releases/download/v0.1.0/kiosk-arm7
sudo cp kiosk-arm7 /usr/bin/kiosk
chmod +x /usr/bin/kiosk
```

**insert and adjust Config yml (see example yml file)**
```
mkdir .kiosk
nano .kiosk/config.yml
```

**disable blank screen and screensaver**
```
sudo nano /etc/xdg/lxsession/LXDE/autostart --> see scripts
```

**install kiosk as service**
- insert given example to ```kiosk.service```
- reload services
- enable service
- start kiosk service

```
sudo nano /etc/systemd/system/kiosk.service 
```
```
sudo systemctl daemon-reload
sudo systemctl enable kiosk
sudo systemctl start kiosk
sudo reboot
```


**Screen resolution**
create ```/usr/share/X11/xorg.conf.d/screen-resolution.conf``` and add
```
Section "Screen"
Identifier "Screen0"
Device "Card0"
SubSection "Display"
Modes "1920x1080"
EndSubSection
EndSection
```

**Screensaver and Screen blanking**
create ```/user/kiosk/xinitrc``` and add
```
xset -dpms
xset s off
```
 