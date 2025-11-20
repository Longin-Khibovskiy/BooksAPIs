package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary
var useCloudinary bool

func InitCloudinary() error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		log.Println("Cloudinary not configured - using local storage for avatars")
		useCloudinary = false

		if err := os.MkdirAll("uploads/avatars", 0755); err != nil {
			return fmt.Errorf("failed to create uploads directory: %w", err)
		}
		return nil
	}

	var err error
	cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		log.Printf("Failed to initialize Cloudinary: %v - falling back to local storage", err)
		useCloudinary = false
		if err := os.MkdirAll("uploads/avatars", 0755); err != nil {
			return fmt.Errorf("failed to create uploads directory: %w", err)
		}
		return nil
	}

	useCloudinary = true
	log.Printf("— Cloudinary initialized successfully (cloud: %s)", cloudName)
	return nil
}

func UploadAvatar(file multipart.File, userID int) (string, error) {
	if useCloudinary && cld != nil {
		return uploadToCloudinary(file, userID)
	}
	return uploadToLocal(file, userID)
}

func uploadToCloudinary(file multipart.File, userID int) (string, error) {
	ctx := context.Background()
	publicID := fmt.Sprintf("user_%d", userID)

	overwrite := true
	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "avatars",
		PublicID:       publicID,
		Transformation: "c_fill,g_center,h_200,w_200",
		Overwrite:      &overwrite,
	})

	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return uploadResult.SecureURL, nil
}

func uploadToLocal(file io.Reader, userID int) (string, error) {
	filename := fmt.Sprintf("user_%d.jpg", userID)
	filePath := filepath.Join("uploads", "avatars", filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return fmt.Sprintf("/uploads/avatars/%s", filename), nil
}
