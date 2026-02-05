package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
	"github.com/chinmay-sawant/gosourcemapper/internal/repository"
	"github.com/chinmay-sawant/gosourcemapper/internal/scanner"
	"github.com/chinmay-sawant/gosourcemapper/internal/scanner/golang"
	"github.com/chinmay-sawant/gosourcemapper/internal/scanner/java"
	"github.com/chinmay-sawant/gosourcemapper/internal/scanner/python"
	"github.com/chinmay-sawant/gosourcemapper/internal/utils"
)

type ScanService interface {
	ScanFile(filePath string, content []byte) ([]*models.CodeNode, error)
	ScanDirectory(dirPath string) ([]*models.CodeNode, error)
	ProcessZipUpload(file *multipart.FileHeader, destRoot string) ([]*models.CodeNode, error)
	GetAllNodes() []*models.CodeNode
}

type scanService struct {
	repo     repository.GraphRepository
	scanners map[string]scanner.Scanner
}

func NewScanService(repo repository.GraphRepository) ScanService {
	// Register scanners
	scanners := make(map[string]scanner.Scanner)
	scanners[".go"] = golang.NewGoScanner()
	scanners[".java"] = java.NewJavaScanner()
	scanners[".py"] = python.NewPythonScanner()

	return &scanService{
		repo:     repo,
		scanners: scanners,
	}
}

func (s *scanService) ScanFile(filePath string, content []byte) ([]*models.CodeNode, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	scn, ok := s.scanners[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	nodes, err := scn.Scan(filePath, content)
	if err != nil {
		return nil, err
	}

	// Save to Repo
	for _, node := range nodes {
		s.repo.SaveNode(node)
	}

	return nodes, nil
}

func (s *scanService) ScanDirectory(dirPath string) ([]*models.CodeNode, error) {
	var allNodes []*models.CodeNode

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories (like .git)
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".temp" { // Allow .temp itself if recursing?
				// Actually standard excludes: .git, .idea via logic
				if info.Name() == ".git" || info.Name() == ".idea" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check extension support
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := s.scanners[ext]; !ok {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		nodes, err := s.ScanFile(path, content)
		if err != nil {
			// Log error but continue? Or fail? Let's log/continue logic basically by collecting err
			// For now, strict fail or ignore?
			// Ignoring error for individual files to allow partial success is better for "mapper"
			return nil
		}
		allNodes = append(allNodes, nodes...)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return allNodes, nil
}

func (s *scanService) ProcessZipUpload(file *multipart.FileHeader, destRoot string) ([]*models.CodeNode, error) {
	// Make sure .temp exists
	if _, err := os.Stat(destRoot); os.IsNotExist(err) {
		if err := os.Mkdir(destRoot, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// Create a unique directory for this upload
	// "temp directory won't be deleted by default"
	folderName := fmt.Sprintf("%s_%d", file.Filename, time.Now().Unix())
	targetDir := filepath.Join(destRoot, folderName)

	if err := os.Mkdir(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Save zip file
	zipPath := filepath.Join(targetDir, file.Filename)
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(zipPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = dst.Close() }()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	// Unzip
	if err := utils.Unzip(zipPath, targetDir); err != nil {
		return nil, err
	}

	// Scan
	return s.ScanDirectory(targetDir)
}

func (s *scanService) GetAllNodes() []*models.CodeNode {
	return s.repo.GetAllNodes()
}
