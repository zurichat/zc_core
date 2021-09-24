package contact

import (
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MAX_FILE_COUNT  = 5
	MAX_FILE_SIZE   = int64(2 * 1024 * 1024) // 2 MB
	MAX_EMAIL_COUNT = 254
)

var (
	EmailRX           = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	AcceptedFileTypes = []string{"jpeg", "jpg", "png", "webp", "bmp", "pdf", "docx", "doc"}
	folderName        = "contact"
	exts              = []string{"png", "jpeg", "jpg", "bmp", "doc", "pdf", "docx"}
)

type Validator struct {
	Errors map[string]string
}

// Create a new Validator instance with an empty errors map
func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if the erros map does not contain any entries
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message to the map as long as no entry already exists fot
// the given key
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds error message to map only if a validation check fails, i.e. not 'ok'
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// In returns true if a specific value is in a list of strings
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Matches returns true if a string value matches a specific regexp pattern
func EmailMatches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Empty returns true if a specific, trimmed value is empty
func Empty(value string) bool {
	return strings.TrimSpace(value) != ""
}

// CharacterCount returns true if the character count of a given value is less
// than a given integer(count)
func CharacterCount(value string, count int) bool {
	return utf8.RuneCountInString(value) < count
}

// AttachmentLength returns true if length of array is less than MAX_FILE_COUNT of 5
func AttachmentLength(list []*multipart.FileHeader) bool {
	return len(list) <= MAX_FILE_COUNT
}

// EmailLength returns true if length of email is less than MAX_EMAIL_COUNT of 254
func EmailLength(email string) bool {
	return len(email) < MAX_EMAIL_COUNT+1
}

// FileSize returns true if the file size of a given file is less than
// MAX_FILE_SIZE of 2MB
func FileSize(file *multipart.FileHeader) bool {
	return file.Size < MAX_FILE_SIZE
}

// AcceptFileSize returns true if the extension of the uploaded files matches the
// allowed file extensions
func AcceptFileType(file *multipart.FileHeader, exts []string) bool {
	filename := file.Filename
	fileExtension := filepath.Ext(filename)
	for _, ext := range exts {
		if ext == fileExtension[1:] {
			return true
		}
	}
	return false
}

// Validate email
func ValidateEmail(validator Validator, email string) {
	validator.Check(EmailLength(email), "email", "email too long")
	validator.Check(EmailMatches(email, EmailRX), "email", "invalid email")
}

// Validate subject
func ValidateSubject(validator Validator, subject string) {
	validator.Check(CharacterCount(subject, 100), "subject", "character count over 100")
	validator.Check(Empty(subject), "subject", "field cannot be empty")
}

// Validate content
func ValidateContent(validator Validator, content string) {
	validator.Check(CharacterCount(content, 500), "content", "character count over 500")
	validator.Check(Empty(content), "content", "field cannot be empty")
}

// Validate attached files
func ValidateAttachedFiles(validator Validator, attachments []*multipart.FileHeader) {
	if len(attachments) > 0 {
		validator.Check(AttachmentLength(attachments), "attachments", "file count exceeded")
		for _, attachment := range attachments {
			validator.Check(AcceptFileType(attachment, exts), "file", "invalid file type")
			validator.Check(FileSize(attachment), "file", "maximum file size exceeded")
		}

	}
}
