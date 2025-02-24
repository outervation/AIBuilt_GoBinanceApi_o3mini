package main

import (
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewRecorder_CreatesNewFile(t *testing.T) {
	instrument := "TEST-INSTR"
	dataType := "testdata"
	batchSize := 10

	type Dummy struct {
		A int `parquet:"name=a, type=INT32"`
	}
	prototype := new(Dummy)

	now := time.Now().UTC()
	expectedDate := now.Format("2006-01-02")
	expectedFilePath := BuildFileName(dataType, instrument, now)

	// Ensure no leftover file exists from previous runs
	if FileExists(expectedFilePath) {
		os.Remove(expectedFilePath)
	}

	r, err := NewRecorder(instrument, dataType, prototype, batchSize)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}

	if r.filePath != expectedFilePath {
		t.Errorf("expected filePath %s, got %s", expectedFilePath, r.filePath)
	}
	if r.currentDate != expectedDate {
		t.Errorf("expected currentDate %s, got %s", expectedDate, r.currentDate)
	}
	if r.localFile == nil {
		t.Error("expected localFile to be initialized, got nil")
	}
	if r.pw == nil {
		t.Error("expected ParquetWriter to be initialized, got nil")
	}

	// Clean up after test
	if err := r.Close(); err != nil {
		t.Errorf("failed to close recorder: %v", err)
	}
	os.Remove(r.filePath)
}

func TestNewRecorder_FileExistsError(t *testing.T) {
	instrument := "TEST-INSTR-EXIST"
	dataType := "testdata"
	batchSize := 10

	type Dummy struct {
		A int `parquet:"name=a, type=INT32"`
	}
	prototype := new(Dummy)

	now := time.Now().UTC()
	filePath := BuildFileName(dataType, instrument, now)

	// Create a dummy file to simulate an existing file.
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}
	f.Close()
	defer os.Remove(filePath)

	// Attempt to create a new recorder, expecting an error due to file existence.
	r, err := NewRecorder(instrument, dataType, prototype, batchSize)
	if err == nil {
		r.Close()
		t.Error("expected error due to existing file, but got nil")
	} else {
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected error message to mention 'already exists', got: %v", err)
		}
	}
}
func TestWriteFlushBatchSize(t *testing.T) {
	instrument := "TEST-INSTR-FLUSH"
	dataType := "testdata"
	batchSize := 3

	type Dummy struct {
		A int `parquet:"name=a, type=INT32"`
	}
	prototype := new(Dummy)

	now := time.Now().UTC()
	filePath := BuildFileName(dataType, instrument, now)
	if FileExists(filePath) {
		os.Remove(filePath)
	}

	r, err := NewRecorder(instrument, dataType, prototype, batchSize)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}

	if len(r.batchBuffer) != 0 {
		t.Errorf("expected empty batchBuffer initially, got %d", len(r.batchBuffer))
	}

	// Write first record; buffer should grow to 1
	err = r.Write(&Dummy{A: 1})
	if err != nil {
		t.Fatalf("unexpected error on first write: %v", err)
	}
	if len(r.batchBuffer) != 1 {
		t.Errorf("expected batchBuffer length 1 after first write, got %d", len(r.batchBuffer))
	}

	// Write second record; buffer should grow to 2
	err = r.Write(&Dummy{A: 2})
	if err != nil {
		t.Fatalf("unexpected error on second write: %v", err)
	}
	if len(r.batchBuffer) != 2 {
		t.Errorf("expected batchBuffer length 2 after second write, got %d", len(r.batchBuffer))
	}

	// Write third record; should trigger flush as batchBuffer reaches batchSize (3 records), and then clear the buffer
	err = r.Write(&Dummy{A: 3})
	if err != nil {
		t.Fatalf("unexpected error on third write (triggering flush): %v", err)
	}
	if len(r.batchBuffer) != 0 {
		t.Errorf("expected batchBuffer to be flushed (length 0) after triggering flush, got %d", len(r.batchBuffer))
	}

	// Write fourth record; should be buffered normally since flush hasn't been triggered again
	err = r.Write(&Dummy{A: 4})
	if err != nil {
		t.Fatalf("unexpected error on fourth write: %v", err)
	}
	if len(r.batchBuffer) != 1 {
		t.Errorf("expected batchBuffer length 1 after fourth write, got %d", len(r.batchBuffer))
	}

	if err := r.Close(); err != nil {
		t.Errorf("failed to close recorder: %v", err)
	}
	os.Remove(r.filePath)
}

func TestRecorder_RotateOnNewDay(t *testing.T) {
	instrument := "TEST-INSTR-ROTATE"
	dataType := "testdata"
	batchSize := 5

	oldNowFunc := NowFunc
	defer func() { NowFunc = oldNowFunc }()

	baseTime := time.Date(2025, 2, 19, 12, 0, 0, 0, time.UTC)
	NowFunc = func() time.Time { return baseTime }

	currentFile := BuildFileName(dataType, instrument, baseTime)
	if FileExists(currentFile) {
		os.Remove(currentFile)
	}
	futureTime := baseTime.Add(24 * time.Hour)
	futureFile := BuildFileName(dataType, instrument, futureTime)
	if FileExists(futureFile) {
		os.Remove(futureFile)
	}

	type Dummy struct {
		A int `parquet:"name=a, type=INT32"`
	}
	prototype := new(Dummy)

	r, err := NewRecorder(instrument, dataType, prototype, batchSize)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	oldFile := r.filePath

	// Write one record
	if err := r.Write(&Dummy{A: 1}); err != nil {
		t.Fatalf("failed to write first record: %v", err)
	}
	if len(r.batchBuffer) != 1 {
		t.Fatalf("expected initial batchBuffer length of 1, got %d", len(r.batchBuffer))
	}

	// Simulate day change: newTime = baseTime + 24 hours
	newTime := baseTime.Add(24 * time.Hour)
	NowFunc = func() time.Time { return newTime }

	expectedNewDate := newTime.Format("2006-01-02")
	expectedNewFile := BuildFileName(dataType, instrument, newTime)

	// Manually trigger rotation with the simulated new time
	if err := r.rotate(newTime); err != nil {
		t.Fatalf("failed to rotate recorder: %v", err)
	}
	// After rotation, the batchBuffer should be empty
	if len(r.batchBuffer) != 0 {
		t.Errorf("expected batchBuffer to be empty after rotation, got %d", len(r.batchBuffer))
	}

	// Write a new record into the new recorder; should be buffered normally
	if err := r.Write(&Dummy{A: 2}); err != nil {
		t.Fatalf("failed to write record after rotation: %v", err)
	}
	if len(r.batchBuffer) != 1 {
		t.Errorf("expected batchBuffer length 1 after writing new record post-rotation, got %d", len(r.batchBuffer))
	}

	// Verify that currentDate and filePath are updated correctly
	if r.currentDate != expectedNewDate {
		t.Errorf("expected currentDate %s, got %s", expectedNewDate, r.currentDate)
	}
	if r.filePath != expectedNewFile {
		t.Errorf("expected new file path %s, got %s", expectedNewFile, r.filePath)
	}

	// Verify that both the old file and the new file exist
	if !FileExists(oldFile) {
		t.Errorf("expected old file %s to exist", oldFile)
	}
	if !FileExists(r.filePath) {
		t.Errorf("expected new file %s to exist", r.filePath)
	}

	// Cleanup: close the recorder and remove both files
	if err := r.Close(); err != nil {
		t.Errorf("failed to close recorder: %v", err)
	}
	os.Remove(oldFile)
	os.Remove(r.filePath)
}
func TestRecorder_CloseFinalizesFile(t *testing.T) {
	instrument := "TEST-INSTR-CLOSE"
	dataType := "testdata"
	batchSize := 5

	type Dummy struct {
		A int `parquet:"name=a, type=INT32"`
	}
	prototype := new(Dummy)
	now := time.Now().UTC()
	filePath := BuildFileName(dataType, instrument, now)
	if FileExists(filePath) {
		os.Remove(filePath)
	}

	r, err := NewRecorder(instrument, dataType, prototype, batchSize)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}

	// Write a couple of records which do not reach the batch size so they remain buffered
	if err := r.Write(&Dummy{A: 100}); err != nil {
		t.Fatalf("failed to write first record: %v", err)
	}
	if err := r.Write(&Dummy{A: 200}); err != nil {
		t.Fatalf("failed to write second record: %v", err)
	}

	// Close the recorder; this should flush the remaining records, finalize the Parquet writer, and close the file
	if err := r.Close(); err != nil {
		t.Fatalf("failed to close recorder: %v", err)
	}

	// Open the file for reading using parquet-go's local file reader
	fr, err := local.NewLocalFileReader(filePath)
	if err != nil {
		t.Fatalf("failed to open parquet file for reading: %v", err)
	}
	pr, err := reader.NewParquetReader(fr, new(Dummy), 4)
	if err != nil {
		t.Fatalf("failed to create ParquetReader: %v", err)
	}

	// Verify that exactly 2 records were written
	num := int(pr.GetNumRows())
	if num != 2 {
		t.Errorf("expected 2 records in file, got %d", num)
	} else {
		records := make([]Dummy, num)
		if err := pr.Read(&records); err != nil {
			t.Errorf("failed to read records: %v", err)
		}
		if len(records) != 2 {
			t.Errorf("expected 2 records, got %d", len(records))
		}
		if records[0].A != 100 || records[1].A != 200 {
			t.Errorf("unexpected record values: %+v", records)
		}
	}
	pr.ReadStop()
	fr.Close()
	os.Remove(filePath)
}
