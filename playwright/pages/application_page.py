class Applications:
    def __init__(self, page):
        self.page = page

    def open_application_page(self):
        self.page.get_by_role("link", name="Applications").click()

    def open_application_details_page(self):
        self.page.get_by_role("link", name="violating-podinfo").first.click()

    def open_application_more_info(self):
        self.page.get_by_text("More Information").click()

    def open_application_events_tab(self):
        self.page.get_by_role("tab", name="Events").click()

    def open_application_graph_tab(self):
        self.page.get_by_role("tab", name="Graph").click()

    def open_application_dependencies_tab(self):
        self.page.get_by_role("tab", name="Dependencies").click()

    def open_application_yaml_tab(self):
        self.page.get_by_role("tab", name="Yaml").click()

    def open_application_violations_tab(self):
        self.page.get_by_role("tab", name="Violations").click()

    def open_application_violations_details(self):
        self.page.get_by_role("link",
                              name="Container Image Pull Policy in deployment violating-podinfo (1 occurrences)"
                              ).nth(0).click()

    def open_policy_details_from_application_violations_details_page(self):
        self.page.get_by_role("link", name="Container Image Pull Policy").click()

    def open_policy_violations_page(self):
        self.page.get_by_role("tab", name="Violations").click()

    def open_policy_violations_details_page(self):
        self.page.get_by_role("link",
                              name="Container Image Pull Policy in deployment violating-podinfo (1 occurrences)"
                              ).nth(0).click()
