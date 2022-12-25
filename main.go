package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Conf struct {
	Dst    string `json:"dst"`
	Src    string `json:"src"`
	Second int    `json:"second"`
}

func confinit() *Conf {

	file, _ := os.Open("conf.json")
	defer file.Close()

	decoder := json.NewDecoder(file)
	conf := Conf{}

	err := decoder.Decode(&conf)

	fmt.Println(conf.Dst)
	if err != nil {
		panic(err)
	}

	return &conf

}

func main() {
	var c *Conf

	c = confinit()

	srcDir := c.Src
	dstDir := c.Dst
	fmt.Println("src :", c.Src+"\n", "Dst :", c.Dst)

	// Daemon uygulamasının çalışma aralığını belirtin (saniye cinsinden)
	tickInterval := time.Duration(c.Second) * time.Second

	// Daemon uygulamasını çalıştır
	for range time.Tick(tickInterval) {
		fmt.Println("hello")
		// Yedeklenecek dizindeki dosya ve dizinlerin listesini al
		entries, err := ioutil.ReadDir(srcDir)
		if err != nil {
			log.Printf("Failed to read source directory: %v", err)
			continue
		}

		// Yedeklenecek dizindeki dosya ve dizinleri tarayın
		for _, entry := range entries {
			srcPath := filepath.Join(srcDir, entry.Name())
			dstPath := filepath.Join(dstDir, entry.Name())

			// Dosya veya dizinin yedeklenmesi gerektiğini kontrol edin
			if !shouldBackup(srcPath, dstPath) {
				continue
			}

			// Dosya veya dizini yedekleyin
			if err := backup(srcPath, dstPath); err != nil {
				log.Printf("Failed to backup %s: %v", srcPath, err)
			} else {
				log.Printf("Successfully backed up %s", srcPath)
			}
		}
	}
}

// shouldBackup dosya veya dizinin yedeklenip yedeklenmeyeceğini kontrol eder.
// Eğer dosya veya dizin yoksa yedeklenir. Eğer dosya veya dizin varsa ve
// son değiştirilme tarihi srcPath'ten sonra dstPath'te değiştirilmişse yedeklenir.
func shouldBackup(srcPath, dstPath string) bool {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false
	}

	dstInfo, err := os.Stat(dstPath)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return false
	}

	return srcInfo.ModTime().After(dstInfo.ModTime())

}

// backup srcPath dosyasını veya dizinini dstPath'e kopyalar.
func backup(srcPath, dstPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		// Eğer kaynak dizin ise, hedef dizin oluşturun ve kaynak dizindeki
		// tüm dosya ve dizinleri hedef dizine kopyalayın
		if err := os.MkdirAll(dstPath, srcInfo.Mode()); err != nil {
			return err
		}

		entries, err := ioutil.ReadDir(srcPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if err := backup(filepath.Join(srcPath, entry.Name()), filepath.Join(dstPath, entry.Name())); err != nil {
				return err
			}
		}
	} else {
		// Eğer kaynak dosya ise, dosyayı hedef dizine kopyalayın
		src, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	return nil
}
