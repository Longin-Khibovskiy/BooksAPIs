package auth

import (
	"fmt"
	"net/http"

	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/services"
)

func UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fmt.Printf("Avatar upload request from user %v\n", userID)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Printf("Error parsing form: %v\n", err)
		http.Redirect(w, r, "/profile?flash=File+too+large", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		fmt.Printf("Error getting file from form: %v\n", err)
		http.Redirect(w, r, "/profile?flash=Error+uploading+file", http.StatusSeeOther)
		return
	}
	defer file.Close()

	fmt.Printf("Received file: %s (size: %d bytes)\n", header.Filename, header.Size)

	avatarURL, err := services.UploadAvatar(file, userID.(int))
	if err != nil {
		fmt.Printf("Error uploading to Cloudinary: %v\n", err)
		http.Redirect(w, r, fmt.Sprintf("/profile?flash=Upload+failed:+%v", err), http.StatusSeeOther)
		return
	}

	_, err = database.DB.Exec(`
		UPDATE users 
		SET avatar_url = $1 
		WHERE id = $2
	`, avatarURL, userID)

	if err != nil {
		fmt.Printf("Error saving avatar URL to database: %v\n", err)
		http.Redirect(w, r, "/profile?flash=Error+saving+avatar", http.StatusSeeOther)
		return
	}

	fmt.Printf("✓ Avatar updated successfully for user %v\n", userID)
	http.Redirect(w, r, "/profile?flash=Avatar+updated+successfully", http.StatusSeeOther)
}
