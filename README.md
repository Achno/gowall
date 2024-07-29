###### *<div align = right><sub>Design By Achno</sub></div>*
<div align = center><img src="assets/go-sticker-overlay-small.png"><br><br>

&ensp;[<kbd> <br> Overview <br> </kbd>](#overview)&ensp;
&ensp;[<kbd> <br> Themes <br> </kbd>](#themes)&ensp;
&ensp;[<kbd> <br> Usage <br> </kbd>](#usage)&ensp;
&ensp;[<kbd> <br> Installation <br> </kbd>](#installation)&ensp;
&ensp;[<kbd> <br> Contributions <br> </kbd>](#contributions)&ensp;
<br><br><br><br></div>


```

 ██████╗  ██████╗ ██╗    ██╗ █████╗ ██╗     ██╗         ██████╗ ██╗   ██╗     █████╗  ██████╗██╗  ██╗███╗   ██╗ ██████╗ 
██╔════╝ ██╔═══██╗██║    ██║██╔══██╗██║     ██║         ██╔══██╗╚██╗ ██╔╝    ██╔══██╗██╔════╝██║  ██║████╗  ██║██╔═══██╗
██║  ███╗██║   ██║██║ █╗ ██║███████║██║     ██║         ██████╔╝ ╚████╔╝     ███████║██║     ███████║██╔██╗ ██║██║   ██║
██║   ██║██║   ██║██║███╗██║██╔══██║██║     ██║         ██╔══██╗  ╚██╔╝      ██╔══██║██║     ██╔══██║██║╚██╗██║██║   ██║
╚██████╔╝╚██████╔╝╚███╔███╔╝██║  ██║███████╗███████╗    ██████╔╝   ██║       ██║  ██║╚██████╗██║  ██║██║ ╚████║╚██████╔╝
 ╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚══════╝    ╚═════╝    ╚═╝       ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ 

```

# Overview :framed_picture:

Gowall is a tool to convert an image ( specifically a wallpaper ) to any color-scheme / pallete you like! 

- It supports `single` and `batch` conversion of images to any of the available themes below.
- It also has the ability to `invert` the colors of the image and convert them later

### Supported formats

`png` `jpeg` `jpg` `webp`

### Planned features

1. `Random mode`: This mode applies a series of transformations to your image in a random order.
For example, it might first invert the colors, then apply the Nord theme, and then switch to the Catppuccin theme, among other possibilities.
The final result will be a unique wallpaper that differs from standard conversions.
Use this mode if you want to explore new color schemes or if you're looking for something beyond typical image conversions.

2. `Extract command`: Using this command on an image will extract the color-scheme and give you a pallete of 5 hex color codes ( aka. pywall )

3. `TUI` : Will also have a pretty TUI version made with `bubbletea`
   
<div align = center><img src="assets/catppuccin.png"><br><br>

<div align = center><img src="assets/everforest.png"><br><br>

<div align = center><img src="assets/invert.png"><br><br> <div>

<div align = left>
  
# Themes :art:

The currently supported themes are featured below, if your favourite theme is missing open an issue or a pull request

- **Catppuccin Mocha**
- **Nord**
- **Everforest**
- **Solarized**
- **Gruvbox**
- **Dracula**
- **Tokyo-moon**
- **Onedark**
- **Monokai**
- **Material**
- **Atom One Light**
- **Sweet**
- **Synthwave 84**
- **Atom Dark**
- **Oceanic Next**
- **Shades of Purple**
- **Arc Dark**
- **Sunset Aurant**
- **Sunset Saffron**
- **Sunset Tangerine**

  <br>

# Usage :gear:


1.  `Singe conversion`

  ```bash
    gowall convert path/to/img.png -t <theme-name-lowercase>
  ```

Notes 🗒️ : 
- `path/to/img.png` does not have to be an absolute path. You can use a relative path with the `~` ex. `~/Pictures/img.png` 
- `<theme-name-lowercase>` is one of the above theme names but in `lowercase` ex. `catppuccin`

<br>

2. `Batch conversion`

   ```bash
     gowall convert -b path/img.png,path/im2.png -t <theme-name-lowercase>
   ```

   ⚠️ Do not leave any white spaces between the comma `,` , do it like this :  `path/img.png,path/im2.png`
<br>

3. `Invert colors`

   ```bash
    gowall invert path/to/img.png
   ```
   You can also batch invert colors with :

   ```bash
    gowall invert -b path/img.png,path/img2.png
   ```
   
<br> 

# Installation :package:

#### Arch linux - AUR

```
yay -S gowall
```

#### Void Linux - XBPS-SRC

Assuming you have [void-packages](https://github.com/void-linux/void-packages)

```bash
git clone https://github.com/elbachir-one/void-templates
cd void-templates/ && cp -r gowall/ void-packages/srcpkgs/
cd void-packages/
./xbps-src pkg gowall
sudo xbps-install -R hostdir/binpkgs gowall
```
#### Build from source

🔨 Clone the repo, build the project and move it inside your `$PATH`

```
git clone https://github.com/Achno/gowall
cd gowall
go build
sudo cp gowall /usr/local/bin/
gowall
```

Notes 🗒️ : You dont have to use `sudo cp gowall /usr/local/bin/` if you have `$GOPATH` setup correctly
Eg. you have the following in your .zshrc / .bashrc
```bash
export GOPATH=$(go env GOPATH)
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```
```bash
go install github.com/Achno/gowall@latest
```

# Contributions :handshake:

If you wish to contribute by adding a new theme please open an `issue`

I would also be very happy if you can provide the `rgb values` of your theme as well :) but not required if it's popular
