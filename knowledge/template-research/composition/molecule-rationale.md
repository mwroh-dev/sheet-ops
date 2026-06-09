# Molecule Rationale

Molecule은 Sheet Ops에서 atom과 organism 사이에 있는 재사용 가능한 workbook
블록이다. Atom은 `append_structured_rows`, `extend_table_formulas`,
`group_summarize`처럼 너무 작아서 템플릿의 업무 의도를 설명하지 못한다.
Organism은 `invoice_line_item_billing`, `attendance_register`처럼 이미 하나의
문서나 업무 산출물에 가깝다. Molecule은 그 사이에서 "이 템플릿들이 같은 문제를
풀고 있다"는 근거를 고정한다.

## 고정 기준

각 molecule은 `molecules.json`에서 다섯 가지 판단 근거를 반드시 가진다.

- 포함 기준: 이 블록으로 분류하려면 어떤 workbook 구조와 업무 반복이 있어야 하는가.
- 제외 기준: 비슷해 보여도 이 블록으로 보면 안 되는 조건은 무엇인가.
- 사용 organism: round-009 roadmap 후보 중 어디에서 이 블록이 반복되는가.
- 필요 atom: 이 블록을 실제 Sheet Ops operation으로 만들 때 필요한 최소 operation 조합은 무엇인가.
- 검증 방법: fixture나 verifier가 이 블록의 성공을 어떤 관찰값으로 판단할 수 있는가.

## 분할 원칙

Molecule은 한 번의 operation보다 커야 하고, 완전한 workbook보다 작아야 한다.
두 개 이상의 organism에서 재사용되면 molecule로 올릴 수 있다. 한 organism에만
강하게 나타나더라도, 별도 검증 문제가 분리된다면 molecule 후보가 될 수 있다.
반대로 단순 색상, 레이블, 장식, 단일 셀 입력처럼 독립 검증 가치가 약한 것은
molecule로 승격하지 않는다.

## 현재 Molecule 근거

- 반복 행 입력 블록: invoice, expense, purchase order, inventory, timesheet,
  ticket, maintenance처럼 같은 열 구조의 행이 계속 늘어나는 organism을 묶는다.
  행 추가 후 수식, 상태, 합계가 함께 따라와야 하므로 단순 `write_values`가 아니다.
- 수식 확장 블록: 새 행, 새 열, 새 기간으로 workbook이 커질 때 수식 범위와 참조가
  유지되어야 하는 문제를 분리한다. 수식 보호와 결과 검증이 함께 필요하다.
- 보호된 요약 영역: 사용자가 입력하는 영역과 Sheet Ops가 계산하는 summary 영역을
  분리한다. 예산, 출석, 성적, 교육, 대출처럼 성공 판정이 summary에 모이는 organism에서
  fixture anchor가 된다.
- 메타데이터 헤더: 고객, 공급자, 기간, 문서 번호, 직원 같은 식별자를 문서형 workbook의
  상단 구조로 분리한다. 제목 텍스트가 아니라 출력과 검증에 연결되는 구조화 필드여야 한다.
- 항목별 합계 블록: line item 행을 subtotal, tax, reimbursable total, payable amount로
  모으는 계산 블록이다. invoice 전용이 아니라 비용, 구매, 공사비에서도 반복된다.
- 출력용 문서 영역: 내부 입력 표와 별개로 검토, 승인, 발송, 출력에 쓰이는 고정 산출물
  영역이다. printable form generation과 보호된 합계 검증이 같이 필요하다.
- 기간 복사 블록: 월, 주, 회계기간, 근무기간처럼 반복되는 기간 구조를 만든다. 새 기간
  label과 날짜는 바뀌고, 입력은 초기화되며, 수식은 보존되어야 한다.
- 이월/마감 요약 블록: 기간 복사만으로 부족한 예산, cash flow, 공사비의 closing,
  opening, variance 연속성을 다룬다. 이전 기간 결과가 다음 기간 시작값에 영향을 준다.
- 매트릭스 성장 블록: 사람 x 날짜, 학생 x 과제, 직원 x 교육항목처럼 행과 열 양쪽으로
  커지는 grid를 다룬다. 일반 row append보다 formula cascade와 header 정규화가 중요하다.
- 점수/완료율 요약 블록: grade, completion rate, rating, proficiency처럼 입력값을
  평가 결과로 바꾸는 블록이다. 단순 count가 아니라 검토 가능한 점수화가 있어야 한다.
- 조회/마스터 연결 블록: SKU, 고객, 직원, rate, 공급자 같은 master 값을 key로 연결한다.
  단순 lookup atom보다 key validation, enrichment, downstream propagation이 함께 필요하다.
- 대조/불일치 판정 블록: 둘 이상의 table을 비교해 matched, missing, mismatched 결과를
  남긴다. lookup과 달리 감사 가능한 mismatch bucket과 summary가 핵심이다.
- 임계값 예외 표시 블록: 재고 부족, 예산 초과, overdue, 위험도처럼 기준을 넘는 행을
  찾아 review workflow로 연결한다. 색상 장식이 아니라 예외 판정이어야 한다.
- 상태 파이프라인 블록: status, owner, priority, due date를 가진 queue와 pipeline을
  다룬다. 상태값 검증과 상태별 summary가 있어야 단순 로그가 아닌 workflow가 된다.
- 일정/타임라인 투영 블록: start/end/date/shift/phase를 grid, roster, timeline view에
  반영한다. 날짜 컬럼만 있는 table이 아니라 행 데이터가 표시 구조로 투영되어야 한다.
- 위험/조치 등록 블록: compliance와 safety처럼 risk/action, severity, owner, due date,
  closure evidence가 함께 필요한 감사형 register를 분리한다. 일반 status queue보다
  review와 closure semantics가 강하다.

## 개편 시 사용법

새 organism을 추가할 때는 먼저 어떤 molecule 조합인지 적는다. 기존 molecule의 포함
기준을 만족하지 않으면 새 molecule을 만들기 전에 제외 기준을 먼저 확인한다. 충돌이
생기면 `used_by_organisms`보다 `include_criteria`, `exclude_criteria`,
`verification_methods`를 우선 판단 근거로 삼는다.
