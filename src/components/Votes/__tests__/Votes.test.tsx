import {fireEvent, render} from "@testing-library/react";
import store from "store";
import {ActionFactory} from "store/action";
import configureStore from "redux-mock-store";
import {Provider} from "react-redux";
import {wrapWithTestBackend} from "react-dnd-test-utils";
import {VoteClientModel} from "types/vote";
import Parse from "parse";
import {Votes} from "components/Votes";

const mockStore = configureStore();

const createVotes = (
  withVotes: boolean,
  activeVoting: boolean,
  className?: string,
  votes: VoteClientModel[] = [
    {
      id: "test-id",
      board: "test-board",
      note: "test-note",
      user: "test-user",
      votingIteration: 1,
    },
  ]
) => {
  const initialState = {
    voteConfiguration: {
      board: "test-board",
      votingIteration: 1,
      voteLimit: 5,
      allowMultipleVotesPerNote: false,
      showVotesOfOtherUsers: false,
    },
  };
  const store = mockStore(initialState);
  const [VoteContext] = wrapWithTestBackend(Votes);
  return (
    <Provider store={store}>
      <VoteContext noteId="test-id" className={className} votes={withVotes ? votes : []} activeVoting={activeVoting} />
    </Provider>
  );
};

describe("Votes", () => {
  describe("should render correctly", () => {
    test("with no votes and disabled voting", () => {
      const votes = render(createVotes(false, false));
      expect(votes.container).toMatchSnapshot();
    });
    test("with no votes and active voting", () => {
      const votes = render(createVotes(false, true));
      expect(votes.container).toMatchSnapshot();
    });
    test("with votes and disabled voting", () => {
      const votes = render(createVotes(true, false));
      expect(votes.container).toMatchSnapshot();
    });
    test("with votes and active voting", () => {
      const votes = render(createVotes(true, true));
      expect(votes.container).toMatchSnapshot();
    });
    test("with additional classname", () => {
      const votes = render(createVotes(false, false, "test-classname"));
      expect(votes.container).toMatchSnapshot();
    });
  });

  describe("should dispatch to store on button press", () => {
    const storeDispatchSpy = jest.spyOn(store, "dispatch");

    test("addVote", () => {
      const votes = render(createVotes(true, true));
      fireEvent.click(votes.container.getElementsByClassName("dot-button")[1]);
      expect(storeDispatchSpy).toHaveBeenCalledWith(ActionFactory.addVote("test-id"));
    });

    test("deleteVote", () => {
      const votes = render(createVotes(true, true));
      fireEvent.click(votes.container.getElementsByClassName("dot-button")[0]);
      expect(storeDispatchSpy).toHaveBeenCalledWith(ActionFactory.deleteVote("test-id"));
    });
  });

  describe("Test allowMultipleVotesPerNote works correctly", () => {
    test("allowMultipleVotesPerNote: false", () => {
      // @ts-ignore
      Parse.User.current = jest.fn(() => ({id: "test-user-2"}));

      const votes = [
        {
          id: "test-vote-0",
          board: "test-board",
          note: "test-id",
          user: "test-user-1",
          votingIteration: 1,
        },
        {
          id: "test-vote-1",
          board: "test-board",
          note: "test-id",
          user: "test-user-2",
          votingIteration: 1,
        },
      ];

      const {container} = render(createVotes(true, true, undefined, votes));

      expect(container.querySelector(".votes")?.firstChild).toHaveClass("dot-button__delete");
      expect(container.querySelector(".votes")?.firstChild).toHaveClass("dot-button--own-vote");
      expect(container.querySelector(".votes")?.childElementCount).toEqual(1);
      expect(container.querySelector(".dot-button__delete")?.firstChild).toHaveClass("dot-button__folded-corner");
      expect((container.querySelector(".dot-button")?.lastChild as HTMLSpanElement).innerHTML).toEqual("2");
    });
  });

  describe("Test voteLimit works correctly", () => {
    test("voteLimit: 0", () => {
      // @ts-ignore
      Parse.User.current = jest.fn(() => ({id: "test-user-2"}));

      const {container} = render(createVotes(false, true));

      expect(container.querySelector(".votes")?.childElementCount).toEqual(1);
      expect(container.querySelector(".votes")?.firstChild).not.toHaveClass("dot-button__delete");
      expect(container.querySelector(".votes")?.firstChild).toHaveClass("dot-button__add");
    });
  });
});