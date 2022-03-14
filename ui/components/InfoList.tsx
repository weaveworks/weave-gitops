import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import Text from "./Text";

export type InfoField = [string, any];

const InfoList = styled(
  ({ items, className }: { className?: string; items: InfoField[] }) => {
    return (
      <table className={className}>
        <tbody>
          {_.map(items, ([k, v]) => (
            <tr key={k}>
              <td>
                <Text capitalize semiBold color="neutral30">
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
  table {
    border-spacing: 0;
  }
  tbody tr td:first-child {
    min-width: 200px;
  }
  tr {
    height: 16px;
  }
`;

export default InfoList;
