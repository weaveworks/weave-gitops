import * as React from "react";
import styled from "styled-components";
import AddGitRepoForm from "../../components/AddGitRepoForm";
import Page from "../../components/Page";

type Props = {
  className?: string;
  appName?: string;
};

function AddGitRepo({ className, appName }: Props) {
  const handleSubmit = (state) => console.log(state);

  return (
    <Page
      title={`Add Git Repository${appName ? ` for ${appName}` : ""}`}
      error={null}
      className={className}
    >
      <AddGitRepoForm onSubmit={handleSubmit} />
    </Page>
  );
}

export default styled(AddGitRepo).attrs({ className: AddGitRepo.name })`
  #repoURL {
    min-width: 800px;
  }
`;
