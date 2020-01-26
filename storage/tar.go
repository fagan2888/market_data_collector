package storage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func GetAllFileName(inputPath, suffix string) []string {
	files := make([]string, 0)
	finfos, err := ioutil.ReadDir(inputPath)
	if err != nil {
		fmt.Printf("err : %s \n", err)
		return files
	}
	for _, fi := range finfos {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(fi.Name(), "csv") {
			files = append(files, fi.Name())
		}
	}
	return files
}

func CompressAllCsv(inputPath, dest string) error {
	csvFiles := make([]*os.File, 0)
	finfos, err := ioutil.ReadDir(inputPath)
	if err != nil {
		fmt.Printf("err : %s \n", err)
		return err
	}

	for _, fi := range finfos {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(fi.Name(), "csv") {
			fmt.Printf("%s\n", fi.Name())

			file, err1 := os.OpenFile(inputPath+"/"+fi.Name(), os.O_RDONLY, 666)
			if err1 != nil {
				fmt.Printf("open %s error:%v", inputPath+fi.Name(), err1)
			}
			csvFiles = append(csvFiles, file)
		}
	}

	return Compress(csvFiles, dest)
}

//压缩 使用gzip压缩成tar.gz
func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := compress(file, "", tw)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, tw *tar.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//解压 tar.gz
func DeCompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}