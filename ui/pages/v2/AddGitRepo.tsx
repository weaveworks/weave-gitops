import * as React from "react";
import styled from "styled-components";
import AddGitRepoForm, {
  GitRepoFormState,
} from "../../components/AddGitRepoForm";
import Page from "../../components/Page";
import useNavigation from "../../hooks/navigation";
import { useCreateRepo } from "../../hooks/sources";
import { V2Routes } from "../../lib/types";

type Props = {
  className?: string;
  appName?: string;
};

function AddGitRepo({ className, appName }: Props) {
  const { navigate } = useNavigation();
  const mutation = useCreateRepo();

  const handleSubmit = (state: GitRepoFormState) => {
    mutation
      .mutateAsync({
        ...state,
        reference: {
          branch: state.branch,
        },
      })
      .then(() =>
        navigate.internal(V2Routes.GitRepo, {
          name: state.name,
          namespace: state.namespace,
        })
      );
  };

  return (
    <Page
      title={`Add Git Repository${appName ? ` for ${appName}` : ""}`}
      error={mutation.error}
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
