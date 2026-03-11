package send_email

type EmailDTO struct {
	Data     EmailData `json:"data"`
	Email    string    `json:"email"`
	Template string    `json:"template"`
}

type EmailData struct {
	Subject    string `json:"subject"`
	PinCode    string `json:"pin_code"`
	SenderMail string `json:"sender_mail"`
}
