package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/otwdev/galaxylib"
)

func main() {
	galaxylib.DefaultGalaxyConfig.InitConfig()

	// for {
	eachFile()

	// 	time.Sleep(1 * time.Minute)
	// }

	//
}

func eachFile() {

	c := make(chan int)

	filepath.Walk("./account", func(path string, info os.FileInfo, err error) error {

		if strings.HasSuffix(path, ".ini") == false {
			return nil
		}

		fmt.Println(err)

		fmt.Println(path)

		//	ali := NewAlihl(path)

		go func(path string) {

			ali := NewAlihl(path)

			for {
				//ali := NewAlihl(path)
				data := ali.GetData()

				c := color.New(color.FgGreen, color.Bold)

				c.Printf("共%d条\n", len(data))

				time.Sleep(30 * time.Second)
			}

		}(path)
		return nil
	})

	<-c
}
