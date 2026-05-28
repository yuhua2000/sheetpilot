package workbook

import (
	"fmt"
	"os"
	"sync"

	"github.com/xuri/excelize/v2"
)

// Manager manages open Excel workbooks.
type Manager struct {
	mu        sync.RWMutex
	workbooks map[string]*Workbook
	counter   int
}

// Workbook represents an open Excel workbook.
type Workbook struct {
	ID   string
	File *excelize.File
	Path string
}

// NewManager creates a new workbook manager.
func NewManager() *Manager {
	return &Manager{
		workbooks: make(map[string]*Workbook),
	}
}

// Open opens an existing Excel file or creates a new one.
func (m *Manager) Open(path string) (*Workbook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	id := fmt.Sprintf("wb_%d", m.counter)

	f, err := excelize.OpenFile(path)
	if err != nil {
		// Create new file if not exists
		f = excelize.NewFile()
		if err := f.SaveAs(path); err != nil {
			return nil, fmt.Errorf("create new file: %w", err)
		}
	}

	wb := &Workbook{
		ID:   id,
		File: f,
		Path: path,
	}
	m.workbooks[id] = wb

	return wb, nil
}

// Save saves the workbook to its current path.
func (m *Manager) Save(id string) error {
	m.mu.RLock()
	wb, ok := m.workbooks[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("workbook %s not found", id)
	}

	return wb.File.Save()
}

// SaveAs saves the workbook to a new path.
func (m *Manager) SaveAs(id, path string) error {
	m.mu.RLock()
	wb, ok := m.workbooks[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("workbook %s not found", id)
	}

	if err := wb.File.SaveAs(path); err != nil {
		return fmt.Errorf("save as: %w", err)
	}

	wb.Path = path
	return nil
}

// Close closes the workbook and removes it from the manager.
func (m *Manager) Close(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	wb, ok := m.workbooks[id]
	if !ok {
		return fmt.Errorf("workbook %s not found", id)
	}

	if err := wb.File.Close(); err != nil {
		return fmt.Errorf("close workbook: %w", err)
	}

	delete(m.workbooks, id)
	return nil
}

// Get returns a workbook by ID.
func (m *Manager) Get(id string) (*Workbook, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wb, ok := m.workbooks[id]
	if !ok {
		return nil, fmt.Errorf("workbook %s not found", id)
	}

	return wb, nil
}

// List returns all open workbooks.
func (m *Manager) List() []*Workbook {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Workbook, 0, len(m.workbooks))
	for _, wb := range m.workbooks {
		list = append(list, wb)
	}

	return list
}

// WorkbookInfo contains metadata about a workbook.
type WorkbookInfo struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	SheetCount int    `json:"sheet_count"`
	FileSize   int64  `json:"file_size"`
}

// GetInfo returns metadata about a workbook.
func (m *Manager) GetInfo(id string) (*WorkbookInfo, error) {
	m.mu.RLock()
	wb, ok := m.workbooks[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("workbook %s not found", id)
	}

	sheets := wb.File.GetSheetList()

	var fileSize int64
	if info, err := os.Stat(wb.Path); err == nil {
		fileSize = info.Size()
	}

	return &WorkbookInfo{
		ID:         wb.ID,
		Path:       wb.Path,
		SheetCount: len(sheets),
		FileSize:   fileSize,
	}, nil
}
