import React from "react";
import styled from "styled-components";
import Flex from "./Flex";
import Icon, { IconType } from "./Icon";
import Link from "./Link";
import Text from "./Text";
export interface Breadcrumb {
  label: string;
  url?: string;
}
interface Props {
  path: Breadcrumb[];
}
export const Breadcrumbs = ({ path = [] }: Props) => {
  return (
    <Flex align>
      {path.map(({ label, url }) => {
        return (
          <Flex align key={label}>
            {url ? (
              <>
                <Link
                  data-testid={`link-${label}`}
                  to={url}
                  textProps={{ bold: true, size: "large", color: "neutral40" }}
                >
                  {label}
                </Link>
                <Icon
                  type={IconType.NavigateNextIcon}
                  size="large"
                  color="neutral40"
                />
              </>
            ) : (
              <Text
                size="large"
                color="neutral40"
                data-testid={`text-${label}`}
              >
                {label}
              </Text>
            )}
          </Flex>
        );
      })}
    </Flex>
  );
};

export default styled(Breadcrumbs)``;
