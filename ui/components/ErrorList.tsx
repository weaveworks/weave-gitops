import { Alert, Box, Collapse } from "@mui/material";
import { sortBy, uniqBy } from "lodash";
import * as React from "react";
import styled from "styled-components";
import Button from "./Button";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Text from "./Text";

interface Error {
  clusterName?: string;
  namespace?: string;
  message?: string;
}

type Props = {
  className?: string;
  errors?: Error[];
};

const BoxWrapper = styled(Box as any)`
  .MuiAlert-root {
    width: auto;
    margin-bottom: ${(props) => props.theme.spacing.base};
    background: ${(props) => props.theme.colors.alertLight};
    border-radius: ${(props) => props.theme.spacing.xs};
  }
  .MuiAlert-action {
    display: inline;
    color: ${(props) => props.theme.colors.alertMedium};
  }
  .MuiIconButton-root:hover {
    background-color: ${(props) => props.theme.colors.alertLight};
  }
  .MuiAlert-icon {
    .MuiSvgIcon-root {
      display: none;
    }
  }
  .MuiAlert-message {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
`;

const ErrorText = styled(Text)`
  margin-left: 8px;
`;

const NavButton = styled(Button)`
  padding: 0;
  min-width: auto;
  margin: 0;
`;

const ErrorsCount = styled.span`
  background: ${(props) => props.theme.colors.alertDark};
  color: ${(props) => props.theme.colors.neutral00};
  padding: 4px;
  border-radius: 4px;
  margin: 0 4px;
`;

function ErrorList({ className, errors }: Props) {
  const [expand, setExpand] = React.useState(true);
  const [index, setIndex] = React.useState(0);

  if (!errors || !errors.length) {
    return null;
  }

  const uniq = uniqBy(errors, (error) =>
    [error.clusterName, error.message].join(),
  );

  const sorted = sortBy(uniq, "clusterName", "namespace", "message");
  const currentError = sorted[index];

  return (
    <BoxWrapper className={className} id="alert-list-errors">
      <Collapse in={expand}>
        <Alert severity="error" onClose={() => setExpand(false)}>
          <Flex align center>
            <Icon size="medium" type={IconType.ErrorIcon} color="alertDark" />
            <ErrorText
              size="medium"
              data-testid="error-message"
              color="neutral40"
            >
              {currentError.clusterName}:&nbsp;
              {currentError.message}
            </ErrorText>
          </Flex>
          <Flex align center>
            <NavButton
              disabled={index === 0}
              data-testid="prevError"
              onClick={() => setIndex((currIndex) => currIndex - 1)}
            >
              <Icon
                type={IconType.NavigateBeforeIcon}
                color="alertMedium"
                size="medium"
              />
            </NavButton>
            <ErrorsCount data-testid="errorsCount">
              {index + 1} / {sorted.length}
            </ErrorsCount>
            <NavButton
              disabled={sorted.length === index + 1}
              id="nextError"
              data-testid="nextError"
              onClick={() => setIndex((currIndex) => currIndex + 1)}
            >
              <Icon
                type={IconType.NavigateNextIcon}
                color="alertMedium"
                size="medium"
              />
            </NavButton>
          </Flex>
        </Alert>
      </Collapse>
    </BoxWrapper>
  );
}

export default styled(ErrorList).attrs({ className: ErrorList.name })`
  width: 100%;
`;
