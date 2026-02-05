package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chinmay-sawant/gosourcemapper/internal/models"
	"github.com/chinmay-sawant/gosourcemapper/internal/repository"
	"github.com/chinmay-sawant/gosourcemapper/internal/resolver"
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
	GetNodes(limit int, nextToken string, skipExts, skipDirs []string) ([]*models.CodeNode, string, error)
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
	var mu sync.Mutex

	// Worker pool settings
	// Limit concurrency to avoid too many open files or contention
	maxWorkers := 20
	jobs := make(chan string, 1000)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				content, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				nodes, err := s.ScanFile(path, content)
				if err == nil && len(nodes) > 0 {
					mu.Lock()
					allNodes = append(allNodes, nodes...)
					mu.Unlock()
				}
			}
		}()
	}

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden directories (like .git)
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." && d.Name() != ".temp" {
				if d.Name() == ".git" || d.Name() == ".idea" || d.Name() == ".vscode" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check extension support
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := s.scanners[ext]; ok {
			jobs <- path
		}
		return nil
	})

	close(jobs)
	wg.Wait()

	if err != nil {
		return nil, err
	}

	// Post-processing: Resolve cross-file dependencies
	depResolver := resolver.NewDependencyResolver()
	depResolver.BuildRegistry(allNodes)
	depResolver.ResolveAll(allNodes)

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

func (s *scanService) GetNodes(limit int, nextToken string, skipExts, skipDirs []string) ([]*models.CodeNode, string, error) {
	offset := 0
	if nextToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(nextToken)
		if err != nil {
			// Only log? or return error?
			// Return error for bad request
			return nil, "", fmt.Errorf("invalid nextToken")
		}
		var decodedOffset int
		_, err = fmt.Sscanf(string(decoded), "%d", &decodedOffset)
		if err == nil {
			offset = decodedOffset
		}
	}

	nodes, nextIndex, err := s.repo.GetNodesPaginated(offset, limit, skipExts, skipDirs)
	if err != nil {
		return nil, "", err
	}

	var newNextToken string

	// Refined Logic:
	// If nodes < limit, we exhausted the list? Yes because we scan until limit or end.
	if len(nodes) == limit {
		newNextToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", nextIndex)))
	}

	return nodes, newNextToken, nil
}
