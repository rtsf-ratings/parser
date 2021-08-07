package fast

import (
    "archive/zip"
    "errors"
    "io"
    "io/ioutil"

    "github.com/rtsf-ratings/parser"
)

type FastParser struct {
}

func readZipFile(reader io.ReaderAt, size int64) ([]byte, error) {
    r, err := zip.NewReader(reader, size)
    if err != nil {
        return nil, err
    }

    for _, f := range r.File {
        if f.Name != "outfrom.xml" {
            continue
        }

        rc, err := f.Open()
        if err != nil {
            return nil, err
        }

        content, err := ioutil.ReadAll(rc)
        if err != nil {
            return nil, err
        }

        rc.Close()
        return content, nil
    }

    return nil, errors.New("outfrom.xml not found")
}

func (parser *FastParser) Parse(reader io.ReaderAt, size int64) (*parser.Result, error) {
    content, err := readZipFile(reader, size)
    if err != nil {
        return nil, err
    }

    return parser.ParseXML(content)
}
