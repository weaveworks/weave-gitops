import os

from playwright.sync_api import Playwright, sync_playwright, expect
import pytest
from pages.login_page import Login

URL = os.getenv("URL")
USER_NAME = os.getenv("USER_NAME")
PASSWORD = os.getenv("PASSWORD")
CLUSTER_NAME = os.getenv("CLUSTER_NAME")


@pytest.fixture(scope="session")
def setup(playwright: Playwright):
    browser = playwright.chromium.launch(slow_mo=1000)
    context = browser.new_context()
    page = context.new_page()
    page.wait_for_load_state("networkidle")
    context.tracing.start(screenshots=True, snapshots=True, sources=True)
    page.goto(URL)
    page.wait_for_load_state("load")
    page.set_default_timeout(5000)
    login_page = Login(page)
    login_page.get_user_name_textbox().click()
    login_page.get_user_name_textbox().fill(USER_NAME)
    page.get_by_placeholder("Username").press("Tab")
    login_page.get_password_textbox().fill(PASSWORD)
    login_page.get_continue_button().click()
    expect(page).to_have_url(f"{URL}/clusters")

    yield context
    context.tracing.stop(path="test-results/execution-tracing.zip")
    page.close()
    context.close()
    browser.close()


@pytest.fixture(scope="class")
def login(setup):
    context = setup
    page = context.new_page()
    page.goto(URL)
    yield page
