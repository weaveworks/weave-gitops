import * as React from "react";
import { Link } from "..";
import {
  Bucket,
  Kustomization,
  HelmRelease,
  HelmRepository,
  GitRepository,
  OCIRepository,
  HelmChart,
} from "../lib/objects";
import Button from "./Button";
import Spacer from "./Spacer";

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

  return hasCreateRequestAnnotation ? (
    <Link to={`/resources/${resource.name}/edit`}>
      <Button>Edit</Button>
    </Link>
  ) : null;
};

const CustomActions = ({ actions }) => {
  return actions?.length > 0
    ? actions?.map((action, index) => (
        <React.Fragment key={index}>
          <Spacer padding="xs" />
          {action}
        </React.Fragment>
      ))
    : null;
};

export default CustomActions;
