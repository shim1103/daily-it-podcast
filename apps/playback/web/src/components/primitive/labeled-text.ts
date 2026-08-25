type LabeledTextProps = {
  tag: keyof HTMLElementTagNameMap;
  datasetKey: string;
  text: string;
};

/**
 * 指定した tag に dataset 属性と textContent をそのまま設定した要素を作る（Primitive Component）。
 *
 * @require datasetKey は camelCase の dataset key
 * @ensure tag 要素へ dataset[datasetKey] = "" と textContent = text を設定して返す
 * @invariant domain 知識を持たない。加工・変換をしない
 */
export function createLabeledText(props: LabeledTextProps): HTMLElement {
  const element = document.createElement(props.tag);
  element.dataset[props.datasetKey] = "";
  element.textContent = props.text;
  return element;
}
