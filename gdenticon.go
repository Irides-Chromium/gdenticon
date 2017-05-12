package main

import (
    "os"
    "flag"
    "fmt"
    "regexp"
    "strconv"
)

// -------------------- START TYPE --------------------

// Point
type Point struct {
    x, y float32
}

func Point_new(x, y float32) Point {
    return Point{x, y}
}

// SVGPath
type SVGPath string

func SVGPath_new() SVGPath {
    return ""
}

func (sp *SVGPath) addPolygon(points []Point) {
    dataString := fmt.Sprintf("M%.1f %.1f", points[0].x, points[0].y)
    for i := 1; i < len(points); i ++ {
        dataString += fmt.Sprintf("L%.1f %.1f", points[i].x, points[i].y)
    }
    *sp += SVGPath(dataString + "Z")
}

func (sp *SVGPath) addCircle(point Point, diameter float32, counterClockwise bool) {
    sweepFlag := 1
    if counterClockwise {
        sweepFlag = 0
    }
    radius := diameter / 2
    *sp += SVGPath(fmt.Sprintf("M%.1f %.1fa%.1f,%.1f 0 1,%.1f %.1f,0a%.1f,%.1f 0 1,%.1f %.1f,0", point.x, point.y + radius, radius, radius, float32(sweepFlag), diameter, radius, radius, float32(sweepFlag), -diameter))
}

// SVGRenderer
type SVGRenderer struct {
    pathsByColor map[string]SVGPath    // [color]path
    path         SVGPath
    size         float32 //size //map[string]float32        // [w/h]length
}

func SVGRenderer_new(size float32) (sr SVGRenderer) {
    sr.pathsByColor = make(map[string]SVGPath)
    sr.path = SVGPath_new()
    sr.size = size //map[string]float32{"w": 256, "h": 256}
    return
}

func (sr *SVGRenderer) beginShape(color string) {
    sr.path = SVGPath_new()
}

func (sr *SVGRenderer) endShape(color string) {
    sr.pathsByColor[color] = sr.path
}

func (sr *SVGRenderer) addPolygon(points []Point) {
    sr.path.addPolygon(points)
}

func (sr *SVGRenderer) addCircle(point Point, diameter float32, counterClockwise bool) {
    sr.path.addCircle(point, diameter, counterClockwise)
}

func (sr *SVGRenderer) toSVG() (svg string) {
    size := int(sr.size)
    // SVG Header
    svg = fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\" preserveAspectRation=\"xMidYMid meet\">", size, size, size, size)

    for color, path := range sr.pathsByColor {
        svg += fmt.Sprintf("<path fill=\"%s\" d=\"%s\"/>", color, path)
    }

    // SVG Footer
    svg += "</svg>"
    return
}

// Transform
type Transform struct {
    x, y, size, rotation float32
}

func Transform_new(x, y, size, rotation float32) Transform {
    return Transform{x, y, size, rotation}
}

func (t *Transform) transformPoint(x, y float32, wh ...float32) (tp Point) {
    right := t.x + t.size
    bottom := t.y + t.size
    var w, h float32
    if len(wh) == 2 {
        w, h = wh[0], wh[1]
    } else {
        w, h = 0, 0
    }

    switch t.rotation {
    case 1:
        tp = Point_new(right - y - h, t.y + x)
    case 2:
        tp = Point_new(right - x - w, bottom - y - h)
    case 3:
        tp = Point_new(t.x + y, bottom - x - w)
    default:
        tp = Point_new(t.x + x, t.y + y)
    }
    return
}

var noTransform = Transform_new(0, 0, 0, 0)

// Graphics
type Graphics struct {
    renderer *SVGRenderer
    transform Transform
}

func Graphics_new(renderer *SVGRenderer) (g Graphics) {
    g.renderer = renderer
    g.transform = noTransform
    return
}

func (g *Graphics) addPolygon(points []float32, invert bool) {
    //if len(points) % 2 == 1 {
    //    points = append(points, Point{0, 0})
    //}

    var di int
    if invert {
        di = -2
    } else {
        di = 2
    }
    transform := g.transform
    transformedPoints := make([]Point, 0)

    i := 0
    if invert {
        i = len(points) - 2
    }
    for ; i < len(points) && i >= 0; i += di {
        transformedPoints = append(transformedPoints, transform.transformPoint(points[i], points[i + 1]))
    }
    g.renderer.addPolygon(transformedPoints)
}

func (g *Graphics) addCircle(x, y, size float32, invert bool) {
    p := g.transform.transformPoint(x, y, size, size)
    g.renderer.addCircle(p, size, invert)
}

func (g *Graphics) addRectangle(x, y, w, h float32, invert bool) {
    g.addPolygon([]float32{x, y,
                           x + w, y,
                           x + w, y + h,
                           x, y + h}, invert)
}

func (g *Graphics) addTriangle(x, y, w, h, r float32, invert bool) {
    points := []float32{x + w, y,
                        x + w, y + h,
                        x, y + h,
                        x, y}
    startind := (int(r) % 4) * 2
    points = append(points[:startind], points[startind + 2:]...)
    g.addPolygon(points, invert)
}

func (g *Graphics) addRhombus(x, y, w, h float32, invert bool) {
    g.addPolygon([]float32{x + w / 2, y,
                           x + w, y + h / 2,
                           x + w / 2, y + h,
                           x, y + h / 2}, invert)
}

// Config
type Config struct {
    saturation float32
    colorLightness, grayscaleLightness func(float32) float32
}

// ------------- END TYPES ----------------

var shapes = map[string][]func(g *Graphics, cell float32, index int){
    "center": {
        func (g *Graphics, cell float32, index int) {
            k := cell * 0.42
            g.addPolygon([]float32{
                0, 0,
                cell, 0,
                cell, cell - k * 2,
                cell - k, cell,
                0, cell}, false)
        },

        func (g *Graphics, cell float32, index int) {
            w := cell * 0.5
            h := cell * 0.8
            g.addTriangle(cell - w, 0, w, h, 2, false)
        },

        func (g *Graphics, cell float32, index int) {
            var s float32 = cell / 3
            g.addRectangle(s, s, cell - s, cell - s, false)
        },

        func (g *Graphics, cell float32, index int) {
            inner := cell * 0.1
            if inner > 1 {
                inner = float32(int(inner))
            } else if inner > 0.5 {
                inner = 1
            }

            var outer float32
            if cell < 6 {
                outer = 1
            } else if cell < 8 {
                outer = 2
            } else {
                outer = float32(int(cell / 4))
            }
            g.addRectangle(outer, outer, cell - inner - outer, cell - inner - outer, false)
        },

        func (g *Graphics, cell float32, index int) {
            m := cell * 0.15
            s := cell * 0.5
            g.addCircle(cell - s - m, cell - s - m, s, false)
        },

        func (g *Graphics, cell float32, index int) {
            inner := float32(int(cell * 0.1))
            outer := float32(int(inner * 4))

            g.addRectangle(0, 0, cell, cell, false)
            g.addPolygon([]float32{
                outer, outer,
                cell - inner, outer,
                outer + (cell - outer - inner) / 2, cell - inner}, true)
        },

        func (g *Graphics, cell float32, index int) {
            g.addPolygon([]float32{
                0, 0,
                cell, 0,
                cell, cell * 0.7,
                cell * 0.4, cell * 0.4,
                cell * 0.7, cell,
                0, cell}, false)
        },

        func (g *Graphics, cell float32, index int) {
            g.addTriangle(cell / 2, cell / 2, cell / 2, cell / 2, 3, false)
        },

        func (g *Graphics, cell float32, index int) {
            g.addRectangle(0, 0, cell, cell / 2, false)
            g.addRectangle(0, cell / 2, cell / 2, cell / 2, false)
            g.addTriangle(cell / 2, cell / 2, cell / 2, cell / 2, 1, false)
        },

        func (g *Graphics, cell float32, index int) {
            inner := cell * 0.14
            if cell > 8 {
                inner = float32(int(inner))
            }

            var outer float32
            if cell < 4 {
                outer = 1
            } else if cell < 6 {
                outer = 2
            } else {
                outer = float32(int(cell * 0.35))
            }
            g.addRectangle(0, 0, cell, cell, false)
            g.addRectangle(outer, outer, cell - outer - inner, cell - outer - inner, true)
        },

        func (g *Graphics, cell float32, index int) {
            inner := cell * 0.12
            outer := inner * 3

            g.addRectangle(0, 0, cell, cell, false)
            g.addCircle(outer, outer, cell - inner - outer, true)
        },

        func (g *Graphics, cell float32, index int) {
            g.addTriangle(cell / 2, cell / 2, cell / 2, cell / 2, 3, false)
        },

        func (g *Graphics, cell float32, index int) {
            m := cell * 0.25
            g.addRectangle(0, 0, cell, cell, false)
            g.addRhombus(m, m, cell - m, cell - m, true)
        },

        func (g *Graphics, cell float32, index int) {
            m := cell * 0.4
            s := cell * 1.2
            if index == 0 {
                g.addCircle(m, m, s, false)
            }
        }},

    "outer": {
        /** @param {Graphics} g */
        func (g *Graphics, cell float32, index int) {
            g.addTriangle(0, 0, cell, cell, 0, false)
        },
        /** @param {Graphics} g */
        func (g *Graphics, cell float32, index int) {
            g.addTriangle(0, cell / 2, cell, cell / 2, 0, false)
        },
        /** @param {Graphics} g */
        func (g *Graphics, cell float32, index int) {
            g.addRhombus(0, 0, cell, cell, false)
        },
        /** @param {Graphics} g */
        func (g *Graphics, cell float32, index int) {
            m := cell / 6
            g.addCircle(m, m, cell - 2 * m, false)
        }}}

func decToHex(v int) (s string) {
    if v < 0 {
        s = "00"
    } else if v < 16 {
        s = "0" + fmt.Sprintf("%x", v)
    } else if v < 256 {
        s = fmt.Sprintf("%x", v)
    } else {
        s = "ff"
    }
    return
}

func hueToRgb(m1, m2, h float32) string {
    if h < 0 {
        h += 6
    } else if h > 6 {
        h -= 6
    }

    var rgb float32
    if h < 1 {
        rgb = m1 + (m2 - m1) * h
    } else if h < 3 {
        rgb = m2
    } else if h < 4 {
        rgb = m1 + (m2 - m1) * (4 - h)
    } else {
        rgb = m1
    }
    return decToHex(int(255 * rgb))
}

func hsl(h, s, l float32) string {
    if s == 0 {
        partialHex := decToHex(int(l * 255))
        return "#" + partialHex + partialHex + partialHex
    } else {
        var m1, m2 float32
        if l <= 0.5 {
            m2 = l * (s + 1)
        } else {
            m2 = l + s - l * s
        }
        m1 = l * 2 - m2
        return "#" + hueToRgb(m1, m2, h * 6 + 2) + hueToRgb(m1, m2, h * 6) + hueToRgb(m1, m2, h * 6 - 2)
    }
}

func correctedHsl(h, s, l float32) string {
    correctors := []float32{0.55, 0.5, 0.5, 0.46, 0.6, 0.55, 0.55}
    corrector := correctors[int(h * 6 + 0.5)]

    if l < 0.5 {
        l = l * corrector * 2
    } else {
        l = corrector + (l - 0.5) * (1 - corrector) * 2
    }

    return hsl(h, s, l)
}

/**
 * Gets a set of identicon color candidates for a specified hue and config.
 */
func colorTheme(hue float32, config Config) []string {
    return []string{
        // Dark gray
        hsl(0, 0, config.grayscaleLightness(0)),
        // Mid color
        correctedHsl(hue, config.saturation, config.colorLightness(0.5)),
        // Light gray
        hsl(0, 0, config.grayscaleLightness(1)),
        // Light color
        correctedHsl(hue, config.saturation, config.colorLightness(1)),
        // Dark color
        correctedHsl(hue, config.saturation, config.colorLightness(0))}
}

func indexof(slice []int, value int) int {
    for p, v := range slice {
        if (v == value) {
            return p
        }
    }
    return -1
}

func getCurrentConfig() Config {
    // Seems no config is available

    var lightness = func(configName string, defaultMin, defaultMax float32) func(float32) float32 {
        return func(value float32) float32 {
            value = defaultMin + value * (defaultMax - defaultMin)
            if value < 0 {
                return 0
            } else if value > 1 {
                return 1
            } else {
                return value
            }
        }
    }

    // saturation, colorLightness, grayscaleLightness
    return Config{0.5, lightness("color", 0.4, 0.8), lightness("grayscale", 0.3, 0.9)}
}


//iconGenerator(renderer, hash, 0, 0, size, 0, getCurrentConfig())
func iconGenerator(renderer *SVGRenderer, hash string, size float32, config Config) {
    if found, _ := regexp.MatchString("(?i)^[0-9a-f]{11,}$", hash); ! found {
        var errstr string
        if _found, _ := regexp.MatchString("(?i)^[0-9a-f]+$", hash); ! _found {
            errstr = "Invalid hash: Not consisted of hex digits."
        } else {
            errstr = "Invalid hash. Too short!"
        }
        panic(errstr)
    }

    padding := 0.08 * size         // Always undefined -> 0.08
    size -= padding * 2
    graphics := Graphics_new(renderer)
    cell := size / 4

    var x, y float32
    x = padding + size / 2 - cell * 2
    y = x

    _hue, _ := strconv.ParseInt(hash[len(hash) - 7:], 16, 0)
    var hue float32 = float32(_hue) / 0xfffffff
    avaliableColors := colorTheme(hue, config)
    color_indexs := make([]int, 0)
    var index int

    isDuplicate := func(values []int, index int) bool {
        if indexof(values, index) >= 0 {
            for i := 0; i < len(values); i ++ {
                if indexof(color_indexs, values[i]) >= 0 {
                    return true
                }
            }
        }
        return false
    }

    for i := 0; i < 3; i ++ {
        _index, _ := strconv.ParseInt(string(hash[8 + i]), 16, 0)
        index = int(_index) % len(avaliableColors)
        if isDuplicate([]int{0, 4}, index) || isDuplicate([]int{2, 3}, index) {
            index = 1
        }
        color_indexs = append(color_indexs, index)
    }

    renderShape := func(colorIndex int, shapes []func(*Graphics, float32, int), index, rota_ind int, coords [][]int) {
        var r int64 = 0
        if rota_ind > 0 {
            r, _ = strconv.ParseInt(string(hash[rota_ind]), 16, 0)
        }
        shape_ind, _ := strconv.ParseInt(string(hash[index]), 16, 0)
        shape := shapes[shape_ind % int64(len(shapes))]
        color := avaliableColors[color_indexs[colorIndex]]
        renderer.beginShape(color)

        for i := 0; i < len(coords); i ++ {
            graphics.transform = Transform_new(x + float32(coords[i][0]) * cell, y + float32(coords[i][1]) * cell, cell, float32(r % 4))
            r ++
            shape(&graphics, cell, i)
        }
        renderer.endShape(color)
    }

    // ACTUAL RENDERING
    // Sides
    renderShape(0, shapes["outer"], 2, 3, [][]int{{1, 0}, {2, 0}, {2, 3}, {1, 3}, {0, 1}, {3, 1}, {3, 2}, {0, 2}})
    // Corners
    renderShape(1, shapes["outer"], 4, 5, [][]int{{0, 0}, {3, 0}, {3, 3}, {0, 3}})
    // Center
    renderShape(2, shapes["center"], 1, 0, [][]int{{1, 1}, {2, 1}, {2, 2}, {1, 2}})
}

func main() {
    var size    int
    var hash    string
    var output  string
    var raw     string
    flag.IntVar(&size, "s", 256, "Size of the output Gdenticon.")
    flag.StringVar(&raw, "r", "", "Use this string and hash it before rendering an icon.")
    flag.Parse()

    defer func() {
        if err := recover(); err != nil {
            fmt.Println(err)
        }
    } ()
    hash = flag.Arg(0)
    output = flag.Arg(1)
    file, err := os.Create(output)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    renderer := SVGRenderer_new(float32(size))
    iconGenerator(&renderer, hash, float32(size), getCurrentConfig())
    file.WriteString(renderer.toSVG())
    file.Sync()
}
