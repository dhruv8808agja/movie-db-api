package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"Simple password", "password123"},
		{"Empty password", ""},
		{"Long password", strings.Repeat("a", 72)}, // bcrypt max is 72 bytes
		{"Special characters", "p@ssw0rd!#$%"},
		{"Unicode password", "пароль密码"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash, "Hash should not equal plain password")

			// Verify the hash is valid bcrypt hash
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			assert.NoError(t, err, "Generated hash should be valid")
		})
	}
}

func TestHashPassword_Randomness(t *testing.T) {
	password := "testpassword"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to salt")
}

func TestHashPassword_TooLong(t *testing.T) {
	// bcrypt has a 72 byte limit
	// Passwords longer than 72 bytes should return an error
	longPassword := strings.Repeat("a", 100)

	hash, err := HashPassword(longPassword)

	assert.Error(t, err, "Password longer than 72 bytes should return error")
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "password length exceeds 72 bytes")
}

func TestCheckPassword_ValidPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, _ := HashPassword(password)

	result := CheckPassword(password, hash)

	assert.True(t, result, "Valid password should return true")
}

func TestCheckPassword_InvalidPassword(t *testing.T) {
	password := "mySecurePassword123"
	wrongPassword := "wrongPassword"
	hash, _ := HashPassword(password)

	result := CheckPassword(wrongPassword, hash)

	assert.False(t, result, "Invalid password should return false")
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, _ := HashPassword(password)

	result := CheckPassword("", hash)

	assert.False(t, result, "Empty password should return false")
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	password := "mySecurePassword123"

	result := CheckPassword(password, "")

	assert.False(t, result, "Empty hash should return false")
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	password := "mySecurePassword123"
	invalidHash := "not-a-valid-bcrypt-hash"

	result := CheckPassword(password, invalidHash)

	assert.False(t, result, "Invalid hash should return false")
}

func TestCheckPassword_CaseSensitive(t *testing.T) {
	password := "MyPassword"
	hash, _ := HashPassword(password)

	// Test with different case
	result1 := CheckPassword("mypassword", hash)
	result2 := CheckPassword("MYPASSWORD", hash)
	result3 := CheckPassword("MyPassword", hash)

	assert.False(t, result1, "Password check should be case-sensitive")
	assert.False(t, result2, "Password check should be case-sensitive")
	assert.True(t, result3, "Exact match should return true")
}

func TestCheckPassword_WithWhitespace(t *testing.T) {
	password := "password123"
	hash, _ := HashPassword(password)

	// Test with trailing/leading whitespace
	result1 := CheckPassword(" password123", hash)
	result2 := CheckPassword("password123 ", hash)
	result3 := CheckPassword(" password123 ", hash)

	assert.False(t, result1, "Password with leading space should not match")
	assert.False(t, result2, "Password with trailing space should not match")
	assert.False(t, result3, "Password with leading and trailing spaces should not match")
}

func TestCheckPassword_SpecialCharacters(t *testing.T) {
	specialPasswords := []string{
		"p@ssw0rd!",
		"test#123$",
		"pass^word&",
		"pwd*()_+",
	}

	for _, password := range specialPasswords {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			assert.NoError(t, err)

			result := CheckPassword(password, hash)
			assert.True(t, result, "Password with special characters should match")
		})
	}
}

func TestCheckPassword_Unicode(t *testing.T) {
	unicodePasswords := []string{
		"пароль",    // Russian
		"密码",       // Chinese
		"パスワード",  // Japanese
		"كلمة",      // Arabic
		"🔒🔑",       // Emojis
	}

	for _, password := range unicodePasswords {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			assert.NoError(t, err)

			result := CheckPassword(password, hash)
			assert.True(t, result, "Unicode password should match")
		})
	}
}

func TestHashPassword_Cost(t *testing.T) {
	password := "testpassword"
	hash, err := HashPassword(password)

	assert.NoError(t, err)

	// Verify that the hash uses the default cost (10)
	cost, err := bcrypt.Cost([]byte(hash))
	assert.NoError(t, err)
	assert.Equal(t, bcrypt.DefaultCost, cost, "Hash should use bcrypt default cost")
}
