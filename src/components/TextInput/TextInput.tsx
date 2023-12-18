import {Children, cloneElement, DetailedHTMLProps, FC, InputHTMLAttributes, ReactElement, ReactNode} from "react";
import classNames from "classnames";
import "./TextInput.scss";

export interface TextInputProps extends DetailedHTMLProps<InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> {
  type?: "text" | "password";
  rightAdornment?: ReactElement;
  leftAdornment?: ReactElement;
  actions?: ReactNode;
  loginType?: "social" | "passkeys";
}

export const TextInput: FC<TextInputProps> = ({className, leftAdornment, rightAdornment, actions, type = "text", loginType = "social", ...other}) => {
  if (!(leftAdornment || rightAdornment || actions)) {
    return <input className={classNames("text-input", className)} type={type} {...other} />;
  }

  if (leftAdornment) {
    const iconProps = Children.only(leftAdornment)?.props;
    leftAdornment = cloneElement(leftAdornment!, {className: classNames(iconProps?.className, "text-input__adornment", "text-input__adornment--left")});
  }

  if (rightAdornment) {
    const iconProps = Children.only(rightAdornment)?.props;
    rightAdornment = cloneElement(rightAdornment!, {className: classNames(iconProps?.className, "text-input__adornment", "text-input__adornment--right")});
  }

  return (
    <div className={classNames({"text-input__container": loginType === "social", "text-input__container-passkeys": loginType === "passkeys"})}>
      <div className="text-input__input-wrapper">
        {leftAdornment}
        <input
          className={classNames("text-input", {"text-input--adornment-left": Boolean(leftAdornment), "text-input--adornment-right": Boolean(rightAdornment)}, className)}
          type={type}
          {...other}
        />
        {rightAdornment}
      </div>
      <div className="text-input__actions">{actions}</div>
    </div>
  );
};
