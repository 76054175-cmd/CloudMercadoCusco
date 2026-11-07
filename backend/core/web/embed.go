// Package web embebe la build estática de apps/web en el binario del core.
//
// En producción el Dockerfile copia apps/web/build a web/dist antes de compilar.
// En desarrollo dist/ solo tiene .gitkeep y la web la sirve Vite (puerto 5173).
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// FS devuelve el contenido de dist/ como raíz.
func FS() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// HasBuild indica si se embebió una build real (existe index.html).
func HasBuild() bool {
	_, err := fs.Stat(FS(), "index.html")
	return err == nil
}
