import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Text from "./Text";

const InfoList = styled(
  ({
    items,
    className,
  }: {
    className?: string;
    items: { [key: string]: any };
  }) => {
    return (
      <table className={className}>
        <tbody>
          {_.map(items, (v, k) => (
            <tr key={k}>
              <td>
                <Text capitalize bold>
                  {k}:
                </Text>
              </td>
              <td>{v || "-"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }
)`
  tbody tr td:first-child {
    min-width: 200px;
  }
  tr {
    height: 16px;
  }
`;

export default InfoList;
