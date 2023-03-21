import * as React from "react";
import styled from "styled-components";
import { Condition } from "../lib/api/core/types.pb";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import { computeMessage, getIndicatorInfo } from "./KubeStatusIndicator";
import { NoDialogDataTable } from "./PodDetail";
import Spacer from "./Spacer";
import { ArrowDropDown, DropDown } from "./SyncButton";
import Text from "./Text";

type StatusProps = {
  conditions: Condition[];
  suspended: boolean;
  showAll?: boolean;
  className?: string;
};

function PageStatus({
  conditions,
  suspended,
  className,
  showAll,
}: StatusProps) {
  let msg = suspended ? "Suspended" : computeMessage(conditions);
  const { icon, color } = getIndicatorInfo(suspended, conditions);
  const [open, setOpen] = React.useState(false);

  return (
    <Flex column>
      <Flex align className={className}>
        <Icon type={icon} color={color} size="medium" />
        <Spacer padding="xs" />
        <Text color="neutral30">{msg}</Text>
        {showAll && (
          <ArrowDropDown variant="outlined" onClick={() => setOpen(!open)}>
            <Icon type={IconType.ArrowDropDownIcon} size="base" />
          </ArrowDropDown>
        )}
      </Flex>
      {showAll && (
        <DropDown open={open}>
          <Flex column>
            <Text bold size="large" color="neutral30">
              Conditions:
            </Text>
            <NoDialogDataTable
              hideSearchAndFilters
              fields={[
                { label: "Type", value: "type" },
                { label: "Status", value: "status" },
                { label: "Reason", value: "reason" },
                { label: "Message", value: "message" },
              ]}
              rows={conditions}
            />
          </Flex>
        </DropDown>
      )}
    </Flex>
  );
}
export default styled(PageStatus).attrs({ className: PageStatus.name })`
  color: ${(props) => props.theme.colors.neutral30};
`;
