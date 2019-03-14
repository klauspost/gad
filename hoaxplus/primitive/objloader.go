package primitive

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// LoadOBJ will load and OBJ and return vertices and deduplicated edges.
func LoadOBJ(b []byte) (verts P3Ds, edges [][2]int) {
	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	known := make(map[uint64]struct{})

	addLine := func(a, b uint64) {
		if a == b {
			return
		}
		if b < a {
			a, b = b, a
		}
		v := a | (b << 32)
		if _, ok := known[v]; ok {
			return
		}
		known[v] = struct{}{}
		edges = append(edges, [2]int{int(a), int(b)})
	}

	for scanner.Scan() {
		t := scanner.Text()
		if strings.HasPrefix(t, "v ") {
			var c Point3D
			n, err := fmt.Sscanf(t, "v %f %f %f", &c.X, &c.Y, &c.Z)
			if err != nil {
				panic(err)
			}
			if n != 3 {
				panic("not 3")
			}
			verts = append(verts, c)
			continue
		}
		if strings.HasPrefix(t, "f ") {
			t := strings.TrimPrefix(t, "f ")
			faces := strings.Fields(t)
			indices := make([]uint64, len(faces))
			for i, v := range faces {
				s, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					panic(err)
				}
				if s == 0 {
					panic("index 0 face found, should start with 1")
				}
				indices[i] = s - 1
			}
			for i, v := range indices {
				if i == 0 {
					// add first to last
					addLine(v, indices[len(indices)-1])
					continue
				}
				addLine(v, indices[i-1])
			}
		}
	}
	return
}
