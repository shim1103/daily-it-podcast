import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { LabeledText, type LabeledTextProps } from "../primitive/labeled-text.tsx";

/**
 * LabeledText を HTMLElement として返す（既存 DOM 組み立て向けの橋渡し）。
 *
 * @require props は LabeledText と同じ契約
 * @ensure LabeledText の静的 markup から得た根 HTMLElement を返す
 * @invariant Feature の本格 JSX 化はしない。橋渡しのみ。React root を持たない
 */
export function mountLabeledText(props: LabeledTextProps): HTMLElement {
  // why: Feature はまだ DOM API で組み立てるため、JSX Primitive を HTMLElement に落とす。
  //   createRoot は unmount なしだと orphan root になるので static markup を使う
  const markup = renderToStaticMarkup(createElement(LabeledText, props));
  const template = document.createElement("template");
  template.innerHTML = markup;
  const element = template.content.firstElementChild;
  /* v8 ignore next 3 -- LabeledText は常に単一 HTML 要素を返す。契約違反時の明示 fail であり通常経路では到達しない */
  if (!(element instanceof HTMLElement)) {
    throw new Error("LabeledText の静的 markup から HTMLElement を得られない");
  }
  return element;
}
