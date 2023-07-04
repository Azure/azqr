// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package renderers

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/Azure/azqr/internal/embeded"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func CreatePBIReport(source string) {
	if runtime.GOOS != "windows" {
		log.Info().Msg("Skipping PowerBI report generation. Since it's only supported on Windows")
		return
	}

	if source == "" {
		log.Fatal().Msg("Please specify the path to the Excel report file")
	}

	xlsx, err := filepath.Abs(source)
	log.Info().Msgf("Generating Power BI dashboard template: %s.pbit", source)
	xlsx = strings.Replace(xlsx, "\\", "\\\\", -1)
	if err != nil {
		panic(err)
	}

	azqrPath := ".azqr"
	if _, err := os.Stat(azqrPath); err == nil {
		err := os.RemoveAll(azqrPath)
		if err != nil {
			panic(err)
		}
	}
	err = os.Mkdir(azqrPath, 0755)
	if err != nil {
		panic(err)
	}
	pbitPath := ".azqr/azqr.pbit"
	if _, err := os.Stat(pbitPath); err == nil {
		err := os.Remove(pbitPath)
		if err != nil {
			panic(err)
		}
	}

	pbit := embeded.GetTemplates("azqr.pbit")
	err = os.WriteFile(pbitPath, []byte(pbit), 0644)
	if err != nil {
		panic(err)
	}

	unzip(pbitPath, ".azqr/output")

	replacePath(".azqr/output/DataModelSchema", xlsx)

	replacePath(".azqr/output/UnappliedChanges", xlsx)

	err = zipFolder(".azqr/output", strings.Replace(source, ".xlsx", ".pbit", -1))
	if err != nil {
		panic(err)
	}

	err = os.RemoveAll(azqrPath)
	if err != nil {
		panic(err)
	}
}

func unzip(source, destination string) {
	archive, err := zip.OpenReader(source)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return
		}
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				panic(err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func zipFolder(source, destination string) error {
	zipfile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if source == path {
				return nil
			}
			path += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.Replace(path[len(source)+1:], "\\", "/", -1)
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return err
	}
	if err = archive.Flush(); err != nil {
		return err
	}
	return nil
}

func readFileUTF16(filename string) ([]byte, error) {
	// Read the file into a []byte:
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Make an tranformer that converts MS-Win default to UTF8:
	win16le := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)

	// Make a transformer that is like win16le, but abides by BOM:
	utf16bom := unicode.BOMOverride(win16le.NewDecoder())

	// Make a Reader that uses utf16bom:
	unicodeReader := transform.NewReader(bytes.NewReader(raw), utf16bom)

	// decode:
	decoded, err := io.ReadAll(unicodeReader)
	return decoded, err
}

func writeFileUTF16(content string) ([]byte, error) {
	e := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()

	var b bytes.Buffer

	unicodeWriter := transform.NewWriter(&b, e)

	_, err := unicodeWriter.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	return b.Bytes(), err
}

func replacePath(source, xlsx string) {
	contentBytes, err := readFileUTF16(source)
	if err != nil {
		panic(err)
	}
	content := string(contentBytes)
	content = strings.Replace(content, "AZQR_REPORT_PATH", xlsx, -1)
	contentBytes, err = writeFileUTF16(content)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(source, contentBytes, 0644)
	if err != nil {
		panic(err)
	}
}
