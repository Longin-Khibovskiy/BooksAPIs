package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

func InitCloudinary() error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" {
		return fmt.Errorf("CLOUDINARY_CLOUD_NAME must be provided")
	}
	if apiKey == "" {
		return fmt.Errorf("CLOUDINARY_API_KEY must be provided")
	}
	if apiSecret == "" {
		return fmt.Errorf("CLOUDINARY_API_SECRET must be provided")
	}

	var err error
	cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	fmt.Printf("✓ Cloudinary initialized successfully (cloud: %s)\n", cloudName)
	return nil
}

func UploadAvatar(file multipart.File, userID int) (string, error) {
	if cld == nil {
		return "", fmt.Errorf("Cloudinary not initialized")
	}

	ctx := context.Background()
	publicID := fmt.Sprintf("user_%d", userID)

	fmt.Printf("Uploading avatar for user %d to Cloudinary...\n", userID)

	overwrite := true
	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "avatars",
		PublicID:       publicID,
		Transformation: "c_fill,g_center,h_200,w_200",
		Overwrite:      &overwrite,
	})

	if err != nil {
		fmt.Printf("✗ Cloudinary upload failed: %v\n", err)
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	fmt.Printf("✓ Avatar uploaded successfully: %s\n", uploadResult.SecureURL)
	return uploadResult.SecureURL, nil
}

func GetDefaultAvatarURL() string {
	if cld == nil {
		return "https://flagstaffatvclub.com/wp-content/uploads/2018/10/Vacancy-1.jpg"
	}

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	if cloudName == "" {
		return "https://flagstaffatvclub.com/wp-content/uploads/2018/10/Vacancy-1.jpg"
	}

	return fmt.Sprintf("https://flagstaffatvclub.com/wp-content/uploads/2018/10/Vacancy-1.jpg")
}
