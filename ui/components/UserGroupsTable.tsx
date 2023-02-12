import * as React from "react";
import { useCallback, useState } from "react";
import styled from "styled-components";
import DataTable, { Field } from "./DataTable";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import MessageBox from "./MessageBox";
import Text from "./Text";

type Props = {
  className?: string;
  rows: string[];
};

const PointerIcon = styled(Icon)`
  cursor: pointer;
  svg {
    fill: ${(props) => props.theme.colors.neutral30};
  }
  &.copied > svg {
    fill: ${(props) => props.theme.colors.primary10};
  }
`;

const CopyToClipboard = ({
  value,
  className,
}: {
  value: string;
  className?: string;
}) => {
  const [copied, setCopied] = useState(false);
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true);
      setTimeout(() => {
        setCopied(false);
      }, 3000);
    });
  }, [value]);

  return (
    <Text onClick={handleCopy} data-testid="github-code" className={className}>
      <span className="CopyText">{value}</span>
      <PointerIcon
        type={copied ? IconType.CheckMark : IconType.FileCopyIcon}
        className={copied ? "copied" : ""}
        size="medium"
      />
    </Text>
  );
};

function UserGroupsTable({ className, rows }: Props) {
  const providerFields: Field[] = [
    {
      label: "Group Name",
      value: (item) => {
        return CopyToClipboard({ value: item, className: "CopyToClipboard" });
      },
    },
  ];

  if (!rows?.length)
    return (
      <Flex wide tall column align>
        <MessageBox>
          <Text size="large" semiBold>
            You are not subscribed to any Group
          </Text>
        </MessageBox>
      </Flex>
    );

  return (
    <DataTable className={className} rows={rows} fields={providerFields} />
  );
}

export default styled(UserGroupsTable).attrs({
  className: UserGroupsTable.name,
})`
  .CopyToClipboard {
    display: inline-flex;
    justify-content: center;
    align-items: center;
  }
  .CopyText {
    margin-right: 8px;
  }
`;
