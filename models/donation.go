package models

type DonantionCampaign struct {
	post_id        string
	donation_id    string
	goal_amount    string
	current_amount string
	deadline       string
}

type DonationTransaction struct {
	transaction_id string
	donation_id    string
	user_id        string
	datetime       string
	amount         string
	status         string
	reference      string
	qr_code_url    string
}
