import * as React from "react";
import { Link } from "..";
import Button from "./Button";
import Spacer from "./Spacer";
import {
  Bucket,
  Kustomization,
  HelmRelease,
  HelmRepository,
  GitRepository,
  OCIRepository,
  HelmChart,
} from "../lib/objects";

type Props = {
  resource:
    | Bucket
    | Kustomization
    | HelmRelease
    | HelmRepository
    | GitRepository
    | OCIRepository
    | HelmChart;
};

export const EditButton = ({ resource }: Props) => {
  const hasCreateRequestAnnotation =
    resource.obj.metadata.annotations?.["templates.weave.works/create-request"];

  return (
    hasCreateRequestAnnotation && (
      <Link to={`/resources/${resource.name}/edit`}>
        <Button>Edit</Button>
      </Link>
    )
  );
};

function CustomActions(actions: any) {
  return actions?.map((action) => (
    <>
      <Spacer padding="xs" />
      {action}
    </>
  ));
}

export default CustomActions;
