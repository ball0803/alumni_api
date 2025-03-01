package models

var MessageEncryptField = []string{
	"Content",
}

var MessageDecryptField = []string{
	"reply_content",
}

var ChatMessageDecryptField = []string{
	"me.message.content",
	"me.message.reply_message_content",
	"other.message.content",
	"other.message.reply_message_content",
}

var CompanyEncryptField = []string{
	"Companies.Position",
	"Companies.SalaryMin",
	"Companies.SalaryMax",
}

var UserEncryptField = []string{
	"StudentInfo.GPAX",
	"StudentInfo.AdmitYear",
	"StudentInfo.GraduateYear",
	"StudentInfo.EducationLevel",
	"StudentInfo.Email",
	"StudentInfo.Github",
	"StudentInfo.Linkedin",
	"StudentInfo.Facebook",
	"StudentInfo.Phone",
	"Companies.Company",
	"Companies.Address",
	"Companies.Position",
}

var StudentInfoDecryptField = []string{
	"gpax",
	"admit_year",
	"graduate_year",
	"education_level",
}

var UserDecryptField = []string{
	"student_info.gpax",
	"student_info.admit_year",
	"student_info.graduate_year",
	"student_info.education_level",
	"contact_info.email",
	"contact_info.github",
	"contact_info.linkedin",
	"contact_info.facebook",
	"contact_info.phone",
	"companies.company",
	"companies.address",
	"companies.position",
}
