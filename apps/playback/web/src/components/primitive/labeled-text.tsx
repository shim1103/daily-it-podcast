import type { ElementType, ReactElement } from "react";

export type LabeledTextProps = {
  tag: keyof HTMLElementTagNameMap;
  datasetKey: string;
  text: string;
};

/**
 * camelCase の dataset key を HTML の data-* 属性名へ写像する（browser dataset と同じ）。
 */
function datasetKeyToDataAttributeName(datasetKey: string): string {
  return `data-${datasetKey.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)}`;
}

/**
 * 指定した tag に dataset 属性と text をそのまま載せる（Primitive Component）。
 *
 * @require datasetKey は camelCase の dataset key
 * @ensure tag 要素へ dataset[datasetKey] = "" と同等の data-* と text を載せた JSX を返す
 * @invariant domain 知識を持たない。text の加工・変換をしない
 */
export function LabeledText({ tag, datasetKey, text }: LabeledTextProps): ReactElement {
  const Tag = tag as ElementType;
  return <Tag {...{ [datasetKeyToDataAttributeName(datasetKey)]: "" }}>{text}</Tag>;
}
