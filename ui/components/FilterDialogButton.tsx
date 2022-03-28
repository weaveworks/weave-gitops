// eslint-disable-next-line
import { ButtonProps } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { Button, Icon, IconType } from "..";

type Props = ButtonProps;

function FilterDialogButton({ className, ...rest }: Props) {
  return (
    <Button {...rest} className={className} variant="text" color="inherit">
      <Icon type={IconType.FilterIcon} size="medium" color="neutral30" />
    </Button>
  );
}

export default styled(FilterDialogButton).attrs({
  className: FilterDialogButton.name,
})``;
