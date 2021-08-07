package parser

import (
    "io"
    "os"
)

type Parser interface {
    Parse(reader io.ReaderAt, size int64) (*Result, error)
}

func ParseFile(parser Parser, filename string) (*Result, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }

    defer file.Close()

    fileinfo, err := file.Stat()
    if err != nil {
        return nil, err
    }

    return parser.Parse(file, fileinfo.Size())
}
