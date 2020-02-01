# gosim
A simple discrete event simulator in go

## Build and Install
```
go get github.com/fengttt/gosim/...
```
The examples use [pixel](https://github.com/faiface/pixel) for
display.  Please follow the instruction to install 
[requirements](https://github.com/faiface/pixel#requirements).
For example, on ubuntu 18LTS, you need to 
```
sudo apt isntall libgl1-meda-dev xorg-dev libglfw3 libglfw3-dev
```

If you are using windows, you need to follow 
[build pixel on windows](https://github.com/faiface/pixel/wiki/Building-Pixel-On-Windows).

## Examples
* conway is a simulator for Conway's Game of Life.   It includes several
interesting inintial configurations.
```
go run examples/conway/conway.go diehard
```
* spread is a very simple model for infectious disease.
```
go run examples/spread/spread.go -n 10 -r 0.08
```

