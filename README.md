###### *<div align = right><sub>Design By Achno</sub></div>*
<div align = center><img src="assets/go-sticker-overlay-small.png"><br><br>

&ensp;[<kbd> <br> Overview <br> </kbd>](#overview-framed_picture)&ensp;
&ensp;[<kbd> <br> Installation <br> </kbd>](#installation-package)&ensp;
&ensp;[<kbd> <br> Contributions <br> </kbd>](#contributions-handshake)&ensp;
<br><br><br><br></div>

---

# Overview :framed_picture:

Gowall is a tool to convert an image ( specifically a wallpaper ) to any color-scheme / palette you like!
It also offers a bunch of image processing features (image to pixel art, Image upscaling, color palette extraction...)

## Docs
Gowall is really well documented with **visual examples**: [Gowall Docs](https://achno.github.io/gowall-docs/)

---

## Features

❗ Although Gowall is a CLI tool, it has an `Image preview` feature that allows
printing images directly in the terminal. See [Gowall Terminal Image preview](https://achno.github.io/gowall-docs/#image-preview)

- Convert Wallpaper's theme 👾 – Recolor an image to match your favorite + (Custom) themes (Catppuccin...).
- AI Image Upscaling 👾- Increase the resolution of the image while preserving or improving its quality.
- Support for Unix pipes/redirection 👾 - Read from stdin and write to stdout.
- Convert Icon's theme 👾 (svg,ico) - Recolor your icons to match a theme.
- Image to pixel art - Transforms your image to the typical blocky appearance of pixel art.
- Replace a specific color in an image - Pretty self explanatory.
- Create a gif from images - Use the images as frames and specify a delay and the number of loops.
- Extract color palette - Extracts all the dominant colors in an image (like pywal).
- Change Image format - Ex. change format from .webp to .png.
- Invert image colors - Pretty self explanatory.
- Draw on the Image - Draw borders,grids on the image
- Remove the background of the image - Pretty self explanatory.
- Effects - Mirror,Flip,Grayscale,change brightness and more to come!
- Daily wallpapers - Explore community-voted wallpapers that reset daily.

---

<div align="center"><img src="assets/custom.png"><br><br></div>

<div align="center"><img src="https://github.com/user-attachments/assets/4029e2b7-b8fd-4738-9334-20a6d01872c7"><br><br></div>

<div align="center"><img src="https://github.com/user-attachments/assets/c9ee11ae-ec99-4371-9daf-e98a66dc9db7"><br><br></div>

<div align="center"><img src="assets/invert.png"><br><br></div>

---

# Themes :art:

You can check the section [here](https://achno.github.io/gowall-docs/themes) on how to create a **Custom Theme**.

The currently supported themes are featured below, if your favourite theme is missing open an issue or a pull request
All themes can be shown (both default and user-created via `~/.config/gowall/config.yml`) by `gowall list`.

- **Catppuccin flavors**
- **Dracula**
- **Everforest**
- **Gruvbox**
- **Nord**
- **Onedark**
- **Solarized**
- **Tokyo-dark/storm/moon**

<details>
  <summary><strong>Click to see more themes</strong></summary>
  <ul>
    <li><strong>Arc Dark</strong></li>
    <li><strong>Atom Dark</strong></li>
    <li><strong>Atom One Light</strong></li>
    <li><strong>Cat Frappe/latte</strong></li>
    <li><strong>Cyberpunk</strong></li>
    <li><strong>Github Light (black & white)</strong></li>
    <li><strong>Kanagawa</strong></li>
    <li><strong>Material</strong></li>
    <li><strong>Melange (Dark & Light)</strong></li>
    <li><strong>Night Owl</strong></li>
    <li><strong>Oceanic Next</strong></li>
    <li><strong>Rose Pine</strong></li>
    <li><strong>Shades of Purple</strong></li>
    <li><strong>Sunset Aurant</strong></li>
    <li><strong>Sunset Saffron</strong></li>
    <li><strong>Sunset Tangerine</strong></li>
    <li><strong>Sweet</strong></li>
    <li><strong>Synthwave 84</strong></li>
  </ul>
</details>

<br>

---

# Installation :package:

Make sure to do `gowall -v` and compare it against the release page version,
since the [docs](https://achno.github.io/gowall-docs/installation) only show the commands/flags and capabilities of the latest released version.

### MacOS (Homebrew)

```sh
brew install gowall
```

Thank you to `chenrui333`. You can find the [ruby formula](https://github.com/Homebrew/homebrew-core/blob/b86ea8e19ae7bf087fab8e2d56cd623eec1e1cf9/Formula/g/gowall.rb) there.

### Arch linux - AUR

```sh
yay -S gowall
```
### Fedora - COPR

```sh
sudo dnf copr enable achno/gowall
sudo dnf install gowall
```

### NixOS - ( Maintainer : [Emily Trau](https://github.com/emilytrau))

```yaml
  environment.systemPackages = [
    pkgs.gowall
  ];
```

More installation options : [here](https://search.nixos.org/packages?channel=24.05&from=0&size=50&sort=relevance&type=packages&query=gowall)

### Void Linux - XBPS-SRC ( Maintainer : [elbachir-one](https://github.com/elbachir-one/))

Assuming you have [void-packages](https://github.com/void-linux/void-packages)

```sh
git clone https://github.com/elbachir-one/void-templates
cd void-templates/ && cp -r gowall/ void-packages/srcpkgs/
cd void-packages/
./xbps-src pkg gowall
sudo xbps-install -R hostdir/binpkgs gowall
```

### Build from source

🔨 Clone the repo, build the project and move it inside your `$PATH`

```sh
git clone https://github.com/Achno/gowall
cd gowall
go build
sudo cp gowall /usr/local/bin/
gowall
```

Notes 🗒️ : You dont have to use `sudo cp gowall /usr/local/bin/` if you have `$GOPATH` setup correctly
Eg. you have the following in your .zshrc / .bashrc

```sh
export GOPATH=$(go env GOPATH)
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

And simply use `go install`

```sh
go install github.com/Achno/gowall@latest
```
---

# Contributions :handshake:

If you wish to contribute by adding a new theme please open an `issue`
I would also be very happy if you can provide the `rgb values` of your theme as well :) but not required if it's popular

Feel free to suggest any cool features that would improve gowall even further by opening an `issue` 

# Special Thanks

Special thanks to [lutgen](https://github.com/ozwaldorf/lutgen-rs) for the original implementation of the color correction algorithm which i adapted for this project.
