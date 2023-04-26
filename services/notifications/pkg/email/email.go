// Package email implements utility for rendering the Email.
//
// The email package supports transifex translation for email templates.
package email

import (
	"bytes"
	"embed"
	"html"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/owncloud/ocis/v2/services/notifications/pkg/channels"
)

var (
	//go:embed templates
	templatesFS embed.FS
)

// RenderEmailTemplate renders the email template for a new share
func RenderEmailTemplate(mt MessageTemplate, locale string, emailTemplatePath string, translationPath string, vars map[string]interface{}) (*channels.Message, error) {
	// translate a message
	mt.Subject = ComposeMessage(mt.Subject, locale, translationPath)
	mt.Greeting = ComposeMessage(mt.Greeting, locale, translationPath)
	mt.MessageBody = ComposeMessage(mt.MessageBody, locale, translationPath)
	mt.CallToAction = ComposeMessage(mt.CallToAction, locale, translationPath)

	// replace the subject email placeholders with the values
	subject, err := executeRaw(mt.Subject, vars)
	if err != nil {
		return nil, err
	}

	// replace the textBody email template placeholders with the translated template
	rawTextBody, err := rowEmailTemplate(emailTemplatePath, mt)
	if err != nil {
		return nil, err
	}
	// replace the textBody email placeholders with the values
	textBody, err := executeRaw(rawTextBody, vars)
	if err != nil {
		return nil, err
	}
	// replace the textBody email template placeholders with the translated template
	mt.Greeting = newlineToBr(mt.Greeting)
	mt.MessageBody = newlineToBr(mt.MessageBody)
	mt.CallToAction = callToActionToHtml(mt.CallToAction)
	rawHtmlBody, err := rowHtmlEmailTemplate(emailTemplatePath, mt)
	if err != nil {
		return nil, err
	}
	// replace the textBody email placeholders with the values
	htmlBody, err := executeRaw(rawHtmlBody, vars)
	if err != nil {
		return nil, err
	}
	return &channels.Message{
		Subject:  subject,
		TextBody: textBody,
		HtmlBody: htmlBody,
	}, nil
}

func rowEmailTemplate(emailTemplatePath string, mt MessageTemplate) (string, error) {
	var err error
	var tpl *template.Template
	// try to lookup the files in the filesystem
	tpl, err = template.ParseFiles(filepath.Join(emailTemplatePath, mt.textTemplate))
	if err != nil {
		// template has not been found in the fs, or path has not been specified => use embed templates
		tpl, err = template.ParseFS(templatesFS, filepath.Join("templates/", mt.textTemplate))
		if err != nil {
			return "", err
		}
	}
	str, err := executeTemplate(tpl, mt)
	if err != nil {
		return "", err
	}
	return html.UnescapeString(str), err
}

func rowHtmlEmailTemplate(emailTemplatePath string, mt MessageTemplate) (string, error) {
	var err error
	var tpl *template.Template

	// try to lookup the files in the filesystem
	tpl, err = template.ParseFiles(filepath.Join(emailTemplatePath, mt.htmlTemplate))
	if err != nil {
		// template has not been found in the fs, or path has not been specified => use embed templates
		_ = filepath.Join("templates/", mt.htmlTemplate)
		tpl, err = template.ParseFS(templatesFS, filepath.Join("templates/", mt.htmlTemplate))
		if err != nil {
			return "", err
		}
	}

	content, err := tpl.ParseFS(templatesFS, filepath.Join("templates", "common", "email.footer.html.tmpl"))
	str, err := executeTemplate(content, mt)
	if err != nil {
		return "", err
	}
	return html.UnescapeString(str), err
}

func executeRaw(raw string, vars map[string]interface{}) (string, error) {
	tpl, err := template.New("").Parse(raw)
	if err != nil {
		return "", err
	}
	return executeTemplate(tpl, vars)
}

func executeTemplate(tpl *template.Template, vars any) (string, error) {
	var writer bytes.Buffer
	if err := tpl.Execute(&writer, vars); err != nil {
		return "", err
	}
	return writer.String(), nil
}

func newlineToBr(s string) string {
	return strings.Replace(s, "\n", "<br>", -1)
}

func callToActionToHtml(s string) string {
	s = strings.TrimSpace(strings.TrimRight(s, "{{ .ShareLink }}"))
	return `<a href="{{ .ShareLink }}">` + s + `</a>`
}
