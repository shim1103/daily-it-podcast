---
name: 契約の置換は旧版と並べて切り、切替はroute・UseCase・webのApiClientを後続で一度に行う。Aは契約とPortまでで、webのApiClientに並置stubを足さない
date: 2026-10-02T17:18:26
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 契約を置き換える時（pull の `since` を `cursor` へ。`2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md`）、A では contracts に新版の schema を旧版と**並べて**切る。Port にも新しい面を、zero の stub で並べて足す。旧版は切替まで動き続ける。
2. **web の ApiClient には A で並置 stub を足さない**。
3. 切替は、後続の実装（C）が route・UseCase・Port・web の ApiClient を**一度に**行い、その時に旧版を消す。
4. 契約が新旧で混在しない、とは言わない。contracts と Port は C まで新旧が並ぶ。並置の削除条件は `todo:` で示す。

## 2. Reason

1. 新版の client は、server の route と UseCase が切り替わるまで実際には使えない。web の同期（Issue #198）は未実装で、client の面だけを先に切っても、それを使う後続は進まない。`lifecycle/split` §2-9 は client の method・型・stub を A に置いてよい場面を挙げるが、その動機は「client 面だけで後続が進む」ことである。本件はその対象外で、衝突しない。
2. web に並置の stub を足すと、`PlaybackApiClient` を返す test double（現状 3 箇所）と、zero を返す stub と、足場 test が増える。これらは C で必ず消す仮物で、置く間は旧版と新版の2系統を web が抱える。
3. 切替を一度に行えば、旧版を消す時点が1回で、「新版は在るが旧版を使い続ける」中間の状態が server と web の間に生じない。route・UseCase・ApiClient のどれかだけが先に新版へ移ると、残りが旧版のまま食い違う。
4. #198 は pull の部分を C の後に回せる。push 側は独立で、先に進められる。A で web を並置しなくても、#198 の進行は止まらない。
5. contracts と Port を A で並べるのは、後続が同じ形を前提に実装できるよう、置換先の境界を先に固定するためである（commit `1bab2dc`、`5117cf2`）。web の ApiClient は、この境界に従う側で、切替の時に合わせれば足りる。

## 3. Rejected

1. **A で web に並置 stub（`pullProgressByCursor` のような新 method）を足し、#198 を先行させる案** — server が切り替わるまで新版の client は使えず、C で消す仮物（test double 3 箇所・zero 返却・足場 test）だけが増える。#198 の push 側は独立に進められるので、pull の stub を急ぐ理由が無い。
2. **A で旧版を消し、新版へ一度に置き換える案** — route・UseCase・web が C まで旧版を使うので、A の時点で消すと実装が壊れる。切替までは旧版を動かし続ける必要がある。
