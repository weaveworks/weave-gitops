import _ from "lodash";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
  pairs: { key: string; value: string | JSX.Element }[];
  columns: number;
};

const Key = styled.div`
  font-weight: bold;
`;

const Value = styled.div``;

function KeyValueTable({ className, pairs, columns }: Props) {
  const arr = new Array(Math.ceil(pairs.length / columns))
    .fill(null)
    .map(() => pairs.splice(0, columns));

  return (
    <div role="list" className={className}>
      <table>
        <tbody>
          {_.map(arr, (a, i) => (
            <tr key={i}>
              {_.map(a, ({ key, value }) => {
                const label = _.capitalize(key);

                return (
                  <td role="listitem" key={key}>
                    <Key aria-label={label}>{label}</Key>
                    <Value>
                      {value || <span style={{ marginLeft: 2 }}>-</span>}
                    </Value>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default styled(KeyValueTable)`
  table {
    width: 100%;
  }

  tr {
    height: 64px;
  }
`;
