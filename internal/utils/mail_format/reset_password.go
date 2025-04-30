package mail_format

const ResetPasswordMail = `
  <!DOCTYPE html>
  <html lang="en">
  <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>Reset Your Password</title>
      <style>
          body {
              font-family: Helvetica, Arial, sans-serif;
              line-height: 1.6;
              color: #333333;
              margin: 0;
              padding: 0;
              background-color: #f0f0f0;
          }
          .container {
              max-width: 600px;
              margin: 0 auto;
              padding: 20px;
          }
          .header {
              text-align: center;
              padding: 20px 0;
          }
          .logo {
              max-width: 150px;
              height: auto;
          }
          .content {
              background-color: #ffffff;
              padding: 30px;
              border-radius: 5px;
              box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
          }
          .button {
              display: block;
              width: 200px;
              margin: 30px auto;
              padding: 12px 0;
              background-color: #1e88e5;
              color: #ffffff;
              text-align: center;
              text-decoration: none;
              font-weight: bold;
              border-radius: 5px;
          }
          .footer {
              margin-top: 30px;
              text-align: center;
              font-size: 12px;
              color: #777777;
          }
          .help-text {
              font-size: 14px;
              color: #555555;
              margin-top: 20px;
          }
      </style>
  </head>
  <body>
      <div class="container">
          <div class="header">
              <img src="c:\Users\CPE\Desktop\CPE-Alumni\cpealumni.png" alt="Customer Portal Logo" class="logo">
          </div>
          <div class="content">
              <h2 style="color: #1e88e5; text-align: center;">Reset Your Password</h2>
              
              <p>Hi [name],</p>
              <p>You recently requested to reset the password for your [CPE Alumni] account. Click the button below to proceed.</p>
              <a href="%s/reset_password?token=%s" class="button">Reset Password</a>
              <p class="help-text">If you did not request a password reset, please ignore this email or reply to let us know. This password reset link is only valid for the next 30 minutes.</p>
              <p>Thanks,<br>the CPE Alumni team</p>
              <h3 style="color: #1e88e5; text-align: center;">Ref: %s</h3>
          </div>
          <div class="footer">
              <p>&copy; 2025 CPE Alumni</p>
              <p>126 Pracha Uthit Rd, Bang Mot, Thung Khru, Bangkok</p>
              <p><a href="#" style="color: #1e88e5;">Privacy Policy</a></p>
          </div>
      </div>
  </body>
  </html>
`
