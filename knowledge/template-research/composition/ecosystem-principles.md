# Advisory Ecosystem Principles

Atom, molecule, organism은 Sheet Ops의 물길을 잡는 hint ecosystem이다. 이것은
runtime 권한 체계가 아니다.

## 원칙

- Advisory ecosystem은 request 해석, fixture 우선순위, verifier 관점, backlog 정렬을 돕는다.
- Runtime support는 capability registry, schema contract, deterministic executor,
  operation-specific verifier가 모두 있을 때만 주장한다.
- Molecule이나 organism에 등장한 planned atom은 capability opportunity일 뿐이다.
- LLM은 ecosystem을 참고할 수 있지만 모든 요청을 억지로 기존 organism에 끼워 맞추면 안 된다.
- Research artifact는 제3자 템플릿의 파일, 수식, 레이아웃, 문구를 재배포하지 않는다.

## 권한 경계

`knowledge/template-research`는 advisory evidence와 planning hint를 담는다.
`contracts/capabilities`와 `runtime`은 supported capability의 권한을 가진다. 이 두
영역이 연결되기 전까지 organism catalog는 제품 기능 claim이 아니다.

## Code-Gen Assist 방향

Code generation assist는 atom builder를 강제하는 것이 아니라 반복 구현 방식을
짧게 떠올리게 하는 scaffolding이다. 좋은 atom builder 기록은 다음 네 가지를 가진다.

- input: 사용자가 주는 값과 workbook에서 탐지해야 하는 값
- plan: 읽기/쓰기 range, formula, validation, protection의 작업 계획
- result: 변경된 workbook 영역과 warning/error
- verification: 성공을 판단할 expected state

이 구조는 토큰을 줄이고 실패 진단을 쉽게 하지만, 아직 구현되지 않은 atom을 supported로
승격시키지는 않는다.
