package main

import (
	"bufio"
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
)

type Snippet struct {
	Code          string `form:"code"`
	TestBenchName string `form:"test_bench_name"`
}

type Result struct {
	Stdout  string `json:"stdout"`
	VcdFile string `json:"vcd_file"`
	Id      string `json:"id"`
}

func runCodeHandler(ctx echo.Context) error {
	cwd, _ := os.Getwd()
	id := RandomString(12)
	os.MkdirAll(path.Join(cwd, id), os.ModePerm)
	var snippet Snippet
	err := ctx.Bind(&snippet)
	if err != nil {
		return err
	}
	codeFile, err := os.Create(path.Join(cwd, id, "code.vhdl"))
	defer codeFile.Close()
	if err != nil {
		return err
	}
	codeFile.WriteString(snippet.Code)
	scriptFile, err := os.Create(path.Join(cwd, id, "run.sh"))
	scriptFile.WriteString(`
#!/bin/bash
ghdl -a code.vhdl
ghdl -e ` + snippet.TestBenchName + `
ghdl -r ` + snippet.TestBenchName + ` --vcd=out.vcd
`)
	cmd := exec.Command("bash", "run.sh")
	cmd.Dir = path.Join(cwd, id)
	out, err := cmd.CombinedOutput()
	f, _ := os.Open(path.Join(cwd, id, "out.vcd"))
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)
	encoded := base64.StdEncoding.EncodeToString(content)
	return ctx.JSON(http.StatusOK, Result{Stdout: string(out), VcdFile: encoded, Id: id})
}

func getVcdHandler(ctx echo.Context) error {
	cwd, _ := os.Getwd()
	return ctx.File(path.Join(cwd, ctx.Param("id"), "out.vcd"))
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Static("/viewer", "vcdrom-trunk/app")
	e.Static("/app", "frontend")

	e.POST("/run", runCodeHandler)
	e.GET("/vcd/:id", getVcdHandler)
	e.GET("/", func(ctx echo.Context) error {
		return ctx.Redirect(http.StatusPermanentRedirect, "/app")
	})
	e.Logger.Fatal(e.Start(":1554"))
}
