class Login:
    def __init__(self, page):
        self.page = page

    def get_user_name_textbox(self):
        return self.page.get_by_placeholder("Username")

    def get_password_textbox(self):
        return self.page.get_by_placeholder("Password")

    def get_continue_button(self):
        return self.page.get_by_role("button", name="CONTINUE")

    def get_account_settings_menu(self):
        return self.page.locator("xpath=//button[@title='Account settings']")

    def get_logout_button(self):
        return self.page.get_by_text("Logout")
