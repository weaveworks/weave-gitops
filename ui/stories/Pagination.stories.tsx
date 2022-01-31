import { Meta, Story } from "@storybook/react";
import React from "react";
import Pagination, { Props } from "../components/Pagination";

export default {
  title: "Pagination",
  component: Pagination,
  parameters: {
    docs: {
      description: {
        component: "Pagination controller for DataTable",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => (
  <Pagination {...args} current={{ start: 0, pageTotal: 0, outOf: 0 }} />
);

export const Default = Template.bind({});
Default.args = {
  className: "",
  forward: () => "",
  skipForward: () => "",
  back: () => "",
  skipBack: () => "",
  perPage: (value) => "",
  current: { start: 0, pageTotal: 0, outOf: 0 },
};
