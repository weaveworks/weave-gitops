import os

from playwright.sync_api import Playwright, sync_playwright, expect
from pages.gitopssets_page import GitopsSets
import pytest


@pytest.mark.usefixtures("login")
class TestGitopsSets:
    @pytest.fixture(autouse=True)
    def _create_obj(self, login):
        self.page = login
        self.gitopssets_page = GitopsSets(self.page)
        self.URL = os.getenv("URL")

    def test_open_gitopssets_page(self):
        self.gitopssets_page.open_gitopssets_page()
        expect(self.page).to_have_url(f"{self.URL}/gitopssets")

    def test_open_gitopssets_details_page(self):
        self.gitopssets_page.open_gitopssets_details_page()
        expect(self.page).to_have_url(f"{self.URL}/gitopssets/object/details?"
                                      "clusterName=management"
                                      "&name=gitopsset-configmaps"
                                      "&namespace=default")

    def test_open_gitopssets_events_tab(self):
        self.gitopssets_page.open_gitopssets_events_tab()
        expect(self.page).to_have_url(f"{self.URL}/gitopssets/object/events?"
                                      "clusterName=management"
                                      "&name=gitopsset-configmaps"
                                      "&namespace=default")

    def test_open_gitopssets_graph_tab(self):
        self.gitopssets_page.open_gitopssets_graph_tab()
        expect(self.page).to_have_url(f"{self.URL}/gitopssets/object/graph?"
                                      "clusterName=management"
                                      "&name=gitopsset-configmaps"
                                      "&namespace=default")

    def test_open_gitopssets_yaml_tab(self):
        self.gitopssets_page.open_gitopssets_yaml_tab()
        expect(self.page).to_have_url(f"{self.URL}/gitopssets/object/yaml?"
                                      "clusterName=management"
                                      "&name=gitopsset-configmaps"
                                      "&namespace=default")
        expect(self.page.get_by_text("kubectl get gitopsset gitopsset-configmaps -n default -o yaml")).to_be_visible()

    def test_back_to_gitopssets_list(self):
        self.gitopssets_page.back_to_gitopssets_list()
        expect(self.page).to_have_url(f"{self.URL}/gitopssets")
        expect(self.page.locator("tbody")).to_contain_text("gitopsset-configmaps")
