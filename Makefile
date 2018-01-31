.DEFAULT: v1tov2
v1tov2:
	@echo "+ $@"
	@docker build -t hinshun/v1tov2 .

.PHONY: v1tov2

