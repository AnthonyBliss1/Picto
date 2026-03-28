<p align="center">
  <img src="build/appicon.png" alt="Picto Icon" width="128" />
</p>

<h1 align="center">Picto</h1>

<p align="center">
  Join / create local rooms and utilize near real-time data transfer via WebSockets
</p>

## Wails
This project was built with Wails V3-Alpha and is required for building

```bash
go install go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

Check dependencies 
```bash
wails3 doctor
```

## Build .app file
This will create an .app file in `Picto/bin`
``` bash 
cd Picto && wails3 build
```

## Dev Testing
```bash
wails3 dev
```

## Note
VPN or advanced router configs can interfere with MDNS service advertisement
