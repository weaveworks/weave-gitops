import * as React from "react";
import styled from "styled-components";
import { Link } from "..";
import Button from "./Button";

type Props = {
  resource: any;
};

function EditButton({ resource }: Props) {
  const hasCreateRequestAnnotation =
    resource.obj.metadata.annotations?.["templates.weave.works/create-request"];

  return (
    hasCreateRequestAnnotation && (
      <Link to={`/resources/${resource.name}/edit`}>
        <Button>Edit</Button>
      </Link>
    )
  );
}

export default EditButton;
