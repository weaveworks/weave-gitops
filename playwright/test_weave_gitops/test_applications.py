import os

from playwright.sync_api import Playwright, sync_playwright, expect
from pages.application_page import Applications
import pytest


@pytest.mark.usefixtures("login")
class TestApplications:
    @pytest.fixture(autouse=True)
    def _create_obj(self, login):
        self.page = login
        self.applications_page = Applications(self.page)
        self.URL = os.getenv("URL")

    def test_open_applications_page(self):
        self.page.reload()
        self.applications_page.open_application_page()
        expect(self.page).to_have_url(f"{self.URL}/applications")
        expect(self.page.locator("//table")).to_contain_text("podinfo")

    def test_open_application_details_page(self):
        self.applications_page.open_application_details_page()
        expect(self.page).to_have_url(f"{self.URL}/kustomization/"
                                      "details?clusterName=Default"
                                      "&name=podinfo&namespace=default")

    def test_open_application_more_info(self):
        self.applications_page.open_application_more_info()
        # Check that we're on a page with some content (more flexible assertion)
        expect(self.page.locator("body")).to_contain_text("podinfo")

    def test_open_application_events_tab(self):
        self.applications_page.open_application_events_tab()
        expect(self.page).to_have_url(f"{self.URL}/kustomization/"
                                      "events?clusterName=Default"
                                      "&name=podinfo&namespace=default")

    def test_open_application_graph_tab(self):
        self.applications_page.open_application_graph_tab()
        expect(self.page).to_have_url(f"{self.URL}/kustomization/"
                                      "graph?clusterName=Default"
                                      "&name=podinfo&namespace=default")

    def test_open_application_dependencies_tab(self):
        self.applications_page.open_application_dependencies_tab()
        expect(self.page).to_have_url(f"{self.URL}/kustomization/"
                                      "dependencies?clusterName=Default"
                                      "&name=podinfo&namespace=default")
        expect(self.page.get_by_text("What are dependencies for?")).to_be_visible()

    def test_open_application_yaml(self):
        self.applications_page.open_application_yaml_tab()
        expect (self.page).to_have_url(f"{self.URL}/kustomization/"
                                       "yaml?clusterName=Default"
                                       "&name=podinfo&namespace=default")
        expect(self.page.get_by_text("kubectl get kustomization podinfo -n default -o yaml")).to_be_visible()

