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
		// TODO
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

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

	if err != nil || len(fileContent) != length {
		// TODO
	}

	return nil
}

func createTar(path, fileName, fileNameNoExt string) error {
	target := fmt.Sprintf("%s.tar", filepath.Join(path, fileNameNoExt))
	tarFile, err := os.Create(target)
	if err != nil {
		// TODO
	}
	defer func(tarFile *os.File) {
		err := tarFile.Close()
		if err != nil {
			// TODO
		}
	}(tarFile)

	info, err := os.Stat(filepath.Join(path, fileName))
	if err != nil {
		// TODO
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		//  TODO
	}

	tarWriter := tar.NewWriter(tarFile)
	err = tarWriter.WriteHeader(header)
	if err != nil {
		// TODO
	}
	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		// TODO
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// TODO
		}
	}(file)

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		// TODO
	}

	return nil
}

func createZip(path, fileName, fileNameNoExt string) error {
	target := fmt.Sprintf("%s.zip", filepath.Join(path, fileNameNoExt))
	zipFile, err := os.Create(target)
	if err != nil {
		// TODO
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {
			// TODO
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)

	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		// TODO
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// TODO
		}
	}(file)

	writer, err := zipWriter.Create(fileName)
	if err != nil {
		// TODO
	}

	if _, err := io.Copy(writer, file); err != nil {
		// TODO
	}

	return nil
}

func main() {

	// should launch for each thread
	tmpDir, err := os.MkdirTemp("", "tmpdir")
	if err != nil {
		// TODO
		return
	}

	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			// TODO
			return
		}
	}(tmpDir)

	fmt.Printf("serving from %s\n", tmpDir)
	if err = os.Chdir(tmpDir); err != nil {
		// TODO
		return
	}

	for _, arch := range architectures() {
		if err = os.Mkdir(arch, 0777); err != nil && os.IsExist(err) {
			// TODO
			return
		}
	}

	// create content
	// change to /usr/share/openshift/LICENSE
	if err = os.Symlink(basePath+"LICENSE", "oc-license"); err != nil {
		// TODO
		return
	}

	var content = []string{"<a href='oc-license'>license</a>"}

	// assign values to 3 variables in for loop from architectures_paths()
	for _, archPath := range architecturesPaths() {
		arch, operatingSystem, pathToArch := archPath[0], archPath[1], archPath[2]
		baseName := path.Base(pathToArch)
		targetPath := filepath.Join(arch, operatingSystem, baseName)
		// targetPath := path.Join(arch, operatingSystem, baseName)
		if err = os.Mkdir(path.Join(tmpDir, arch, operatingSystem), 0777); err != nil {
			// TODO
			return
		}

		if err = os.Symlink(pathToArch, path.Join(tmpDir, targetPath)); err != nil {
			// TODO
			return
		}

		noExtName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		archivePathRoot := path.Join(tmpDir, arch, operatingSystem)
		if err = createTar(archivePathRoot, baseName, noExtName); err != nil {
			// TODO
			return
		}

		if err = createZip(archivePathRoot, baseName, noExtName); err != nil {
			// TODO
			return
		}
		content = append(content, fmt.Sprintf("<a href=\"%s\">oc (%s %s)</a> (<a href=\"%s.tar\">tar</a> <a href=\"%s.zip\">zip</a>)",
			targetPath,
			arch,
			operatingSystem,
			filepath.Join(arch, operatingSystem, noExtName),
			filepath.Join(arch, operatingSystem, noExtName),
		))
	}

	tmpDirBase := filepath.Base(tmpDir)

	//writeAllIndexes(tmpDir, tmpDir)

	if err = filepath.WalkDir(tmpDir+"/", func(path string, d os.DirEntry, err error) error {
		// skip tmpDir
		if d.Name() != tmpDirBase {
			// put listings in each subdirectory
			if d.IsDir() {

				if err = writeIndex(
					filepath.Join(path, "index.html"),
					fmt.Sprintf("<p>Directory listings are disabled. See <a href=\"/\">here</a> for available content.</p>")); err != nil {
					// TODO
				}
			}
		} else {
			formattedContent := []string{"<ul>"}
			for _, c := range content {
				formattedContent = append(formattedContent, fmt.Sprintf("  <li>%s</li>", c))
			}
			formattedContent = append(formattedContent, "</ul>")
			indexFileContent := strings.Join(formattedContent[:], "\n")
			if err = writeIndex(filepath.Join(path, "index.html"), indexFileContent); err != nil {
				// TODO
			}
		}

		return nil
	}); err != nil {
		// TODO
	}

	server := &http.Server{Addr: ":3333", Handler: http.FileServer(http.Dir(tmpDir))}
	server.SetKeepAlivesEnabled(false)
	fmt.Println("listening on port: 3333")

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
