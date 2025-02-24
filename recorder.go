package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"
	"github.com/xitongsys/parquet-go/parquet"
)

var NowFunc = time.Now

// Recorder encapsulates a parquet-go writer and a local file handle.
// It enforces a naming convention (one file per instrument per UTC date with data type in the filename),
// checks for existing files to prevent resuming, rotates files when a new UTC day starts, and batches writes
// to minimize dynamic allocations.
// This implementation follows a functional core, imperative shell approach to facilitate unit testing.

type Recorder struct {
	instrument  string
	dataType    string
	batchSize   int
	currentDate string
	filePath    string
	localFile   *local.LocalFile
	pw          *writer.ParquetWriter
	batchBuffer []interface{}
	prototype   interface{}
}

// NewRecorder creates a new Recorder for the given instrument and data type using the provided prototype
// (which defines the parquet schema) and batchSize. It builds the file name based on the current UTC date,
// and returns an error if a file for the current day already exists (to avoid resuming).
func NewRecorder(instrument string, dataType string, prototype interface{}, batchSize int) (*Recorder, error) {
	now := NowFunc().UTC()
	currentDate := now.Format("2006-01-02")
	fileName := BuildFileName(dataType, instrument, now)
	if FileExists(fileName) {
		return nil, fmt.Errorf("file %s already exists, not resuming recording", fileName)
	}

	lf, err := local.NewLocalFileWriter(fileName)
	if err != nil {
		return nil, err
	}
	lfConcrete, ok := lf.(*local.LocalFile)
	if !ok {
		lf.Close()
		return nil, fmt.Errorf("failed type assertion for local file")
	}

	pw, err := writer.NewParquetWriter(lf, prototype, int64(batchSize))
	if err != nil {
		lf.Close()
		return nil, err
	}

	pw.RowGroupSize = 128 * 1024 * 1024 // 128 MB
	pw.PageSize = 8 * 1024             // 8 KB
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	return &Recorder{
		instrument:  instrument,
		dataType:    dataType,
		batchSize:   batchSize,
		currentDate: currentDate,
		filePath:    fileName,
		localFile:   lfConcrete,
		pw:          pw,
		batchBuffer: make([]interface{}, 0, batchSize),
		prototype:   prototype,
	}, nil
}

// Write adds a record to the Recorder. It performs file rotation if the current UTC day has changed and batches
// the writes. Once the batch size is reached, the buffered records are flushed to the parquet writer.
func (r *Recorder) Write(record interface{}) error {
	now := NowFunc().UTC()
	currentDay := now.Format("2006-01-02")
	if currentDay != r.currentDate {
		if err := r.rotate(now); err != nil {
			return err
		}
	}

	r.batchBuffer = append(r.batchBuffer, record)
	if len(r.batchBuffer) >= r.batchSize {
		return r.flushBuffer()
	}
	return nil
}

// flushBuffer writes all buffered records to the parquet writer and then resets the buffer.
func (r *Recorder) flushBuffer() error {
	for _, rec := range r.batchBuffer {
		if err := r.pw.Write(rec); err != nil {
			return err
		}
	}
	r.batchBuffer = r.batchBuffer[:0]
	return nil
}

// rotate finalizes the current file and starts a new parquet file for the new day.
func (r *Recorder) rotate(newTime time.Time) error {
	if err := r.flushBuffer(); err != nil {
		return err
	}
	if err := r.pw.WriteStop(); err != nil {
		return err
	}
	if err := r.localFile.Close(); err != nil {
		return err
	}

	newDate := newTime.Format("2006-01-02")
	newFileName := BuildFileName(r.dataType, r.instrument, newTime)
	if FileExists(newFileName) {
		return errors.New(fmt.Sprintf("file %s already exists, not resuming recording", newFileName))
	}

	lf, err := local.NewLocalFileWriter(newFileName)
	if err != nil {
		return err
	}
	pw, err := writer.NewParquetWriter(lf, r.prototype, int64(r.batchSize))
	if err != nil {
		lf.Close()
		return err
	}
	pw.RowGroupSize = 128 * 1024 * 1024 // 128 MB
	pw.PageSize = 8 * 1024             // 8 KB
	pw.CompressionType = parquet.CompressionCodec_SNAPPY
	
	lfConcrete, ok := lf.(*local.LocalFile)
	if !ok {
		lf.Close()
		return fmt.Errorf("failed type assertion for local file in rotate")
	}

	r.localFile = lfConcrete
	r.currentDate = newDate
	r.pw = pw
	r.filePath = newFileName
	r.batchBuffer = r.batchBuffer[:0]
	return nil
}

// Close flushes any remaining buffered records, finalizes the parquet writer, and closes the underlying file.
func (r *Recorder) Close() error {
	if err := r.flushBuffer(); err != nil {
		return err
	}
	if err := r.pw.WriteStop(); err != nil {
		return err
	}
	return r.localFile.Close()
}
