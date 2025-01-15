import os

from playwright.sync_api import Playwright, sync_playwright, expect
from pages.policies_page import Policies
import pytest


@pytest.mark.usefixtures("login")
class TestPolicies:

    @pytest.fixture(autouse=True)
    def _obj(self, login):
        self.page = login
        self.policies_page = Policies(self.page)
        self.URL = os.getenv("URL")

    def test_open_policies_page(self):
        self.policies_page.open_policies_page()
        expect(self.page).to_have_url(f"{self.URL}/policies/list")

    def test_open_policy_details_page(self):
        self.policies_page.open_policy_details_page()
        expect(self.page).to_have_url(f"{self.URL}/policy_details/"
                                      "details?clusterName=management"
                                      "&id=weave.policies.containers-minimum-replica-count"
                                      "&name=Containers%20Minimum%20Replica%20Count")
