package main

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

const basePath string = "/Users/pakratoc/GolandProjects/awesomeProject/sources/"

func architectures() []string {
	return []string{"amd64", "arm64", "ppc64le", "s390x"}
}

func architecturesPaths() [][]string {
	// constant array
	// change to /usr/share/openshift/linux_amd64/oc
	return [][]string{
		{"amd64", "linux", basePath + "linux_amd64/oc"},
		{"amd64", "mac", basePath + "mac/oc"},
		{"amd64", "windows", basePath + "windows/oc.exe"},
		{"arm64", "linux", basePath + "linux_arm64/oc"},
		{"arm64", "mac", basePath + "mac_arm64/oc"},
		{"ppc64le", "linux", basePath + "linux_ppc64le/oc"},
		{"s390x", "linux", basePath + "linux_s390x/oc"},
	}
}

func writeIndex(path, message string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	defer closeFile(file)

	fileContentSlice := []string{
		"<!doctype html>",
		"<html lang=\"en\">",
		"<head>",
		"<meta charset=\"utf-8\">",
		"</head>",
		"<body>",
		fmt.Sprintf("%s", message),
		"</body>",
		"</html>",
	}
	fileContent := strings.Join(fileContentSlice[:], "\n")

	length, err := file.WriteString(fileContent)

	if err != nil {
		log.Fatal(err)
	}
	if len(fileContent) != length {
		log.Fatal("Error writing to file")
	}

	return nil
}

func createTar(path, fileName, fileNameNoExt string) {
	tarFile, err := os.Create(fmt.Sprintf("%s.tar", filepath.Join(path, fileNameNoExt)))
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(tarFile)

	info, err := os.Stat(filepath.Join(path, fileName))
	if err != nil {
		log.Fatal(err)
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		log.Fatal(err)
	}

	tarWriter := tar.NewWriter(tarFile)
	err = tarWriter.WriteHeader(header)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)

	if _, err = io.Copy(tarWriter, file); err != nil {
		log.Fatal(err)
	}
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func createZip(path, fileName, fileNameNoExt string) {
	target := fmt.Sprintf("%s.zip", filepath.Join(path, fileNameNoExt))
	zipFile, err := os.Create(target)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(zipFile)

	zipWriter := zip.NewWriter(zipFile)

	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)

	writer, err := zipWriter.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(writer, file); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// should launch for each thread
	tmpDir, err := os.MkdirTemp("", "tmpdir")
	if err != nil {
		log.Fatal(err)
	}

	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Fatal(err)
		}
	}(tmpDir)

	if err = os.Chdir(tmpDir); err != nil {
		log.Fatal(err)
	}

	for _, arch := range architectures() {
		// TODO: correct permission?
		if err = os.Mkdir(arch, 0777); err != nil && os.IsExist(err) {
			log.Fatal(err)
		}
	}

	// create content
	// change to /usr/share/openshift/LICENSE
	if err = os.Symlink(basePath+"LICENSE", "oc-license"); err != nil {
		log.Fatal(err)
	}

	var content = "<ul>\n<li>\n<a href=\"oc-license\">license</a>\n</li>\n"

	// assign values to 3 variables in for loop from architectures_paths()
	for _, archPath := range architecturesPaths() {
		arch, operatingSystem, pathToArch := archPath[0], archPath[1], archPath[2]
		baseName := path.Base(pathToArch)
		noExtName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		archivePathRoot := path.Join(tmpDir, arch, operatingSystem)
		targetPath := filepath.Join(arch, operatingSystem, baseName)

		if err = os.Mkdir(path.Join(tmpDir, arch, operatingSystem), 0777); err != nil {
			log.Fatal(err)
		}

		if err = os.Symlink(pathToArch, path.Join(tmpDir, targetPath)); err != nil {
			log.Fatal(err)
		}

		createTar(archivePathRoot, baseName, noExtName)
		createZip(archivePathRoot, baseName, noExtName)

		content += fmt.Sprintf("<li><a href=\"%s\">oc (%s %s)</a> (<a href=\"%s.tar\">tar</a> <a href=\"%s.zip\">zip</a>)</li>\n",
			targetPath,
			arch,
			operatingSystem,
			filepath.Join(arch, operatingSystem, noExtName),
			filepath.Join(arch, operatingSystem, noExtName),
		)
	}
	content += "</ul>\n"

	if err = filepath.WalkDir(tmpDir+"/", func(path string, d os.DirEntry, err error) error {
		// create index files in subdirectories
		if d.Name() != filepath.Base(tmpDir) {
			// put indexes in each subdirectory
			if d.IsDir() {
				subdirectoryIndexMsg := "<p>Directory listings are disabled. See <a href=\"/\">here</a> for available content.</p>"
				if err = writeIndex(filepath.Join(path, "index.html"), fmt.Sprintf(subdirectoryIndexMsg)); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			// create main index file in root folder
			if err = writeIndex(filepath.Join(path, "index.html"), content); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	server := &http.Server{Addr: ":8080", Handler: http.FileServer(http.Dir(tmpDir))}
	server.SetKeepAlivesEnabled(false)
	fmt.Println("listening on port: 8080")
	log.Fatal(server.ListenAndServe())
}

func init() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()
}
