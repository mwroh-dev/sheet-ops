# Organism Rationale

Organism은 하나의 업무 문서나 반복 가능한 workbook 산출물에 가까운 단위다.
하지만 이 catalog의 organism은 실행 권한이 아니다. 각 항목은 request compiler,
reviewer, fixture planner가 참고할 수 있는 advisory 조합이다.

## 분류 기준

Organism은 다음 조건을 만족할 때 둔다.

- 사용자 workflow가 molecule 하나로 설명되지 않는다.
- 최소 두 개 이상의 molecule이 결합되어 업무 산출물이 된다.
- verifier가 확인해야 할 관찰 지점이 명확하다.
- round-009 `roadmap_candidate`로 판단된 pattern과 일치한다.

반대로 runtime support는 organism catalog가 아니라 capability registry, schema,
runtime executor, verifier fixture가 모두 갖춰졌을 때만 주장한다.

## 현재 판단

- 청구서, 경비 정산, 구매 주문은 문서형 organism이다. 세 항목 모두 metadata,
  line item, total, printable output을 공유하지만, 구매 주문은 상태 파이프라인이
  붙어 별도 organism으로 둔다.
- 예산과 cash flow는 기간형 organism이다. 둘 다 period copy와 carry-forward가
  필요하지만, 예산은 threshold exception이 더 강하고 cash flow는 balance continuity가
  핵심이다.
- 출석부, 근무표, 성적부, 교육 matrix는 matrix 계열이다. 하지만 출석은 기간 grid,
  근무표는 배정 coverage, 성적은 score 계산, 교육은 completion review가 중심이라
  verifier가 달라진다.
- inventory, procurement, warehouse reorder는 운영 데이터 계열이다. inventory는
  movement log, procurement는 reconciliation, reorder는 threshold exception과
  master lookup이 primary다.
- ticket, sales, maintenance, compliance, safety는 workflow/status 계열이다.
  status pipeline을 공유하지만 risk/action semantics나 printable review output의
  강도에 따라 나눈다.
- loan repayment는 일반 period roll-forward가 아니라 calculation schedule이다.
  이 때문에 `calculation_schedule_block`을 별도 molecule로 추가했다.

## 충돌 해결

새로운 template 후보가 기존 organism과 비슷하면 먼저 molecule 조합을 비교한다. 조합이
같고 verifier focus도 같으면 새 organism이 아니라 domain variant로 둔다. 조합이
비슷해도 verifier focus가 달라지면 organism으로 분리할 수 있다.
