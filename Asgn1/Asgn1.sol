// SPDX-License-Identifier: GPL-3.0
pragma solidity  ^0.8.18;

contract TicketBooking{

    // State Variables

    struct Buyer{
        uint totalPrice;
        uint numTickets;
        string email;
    }
    
    address public seller;
    uint public numTicketsSold;
    uint public maxOccupancy;
    uint public price;
    mapping(address => Buyer) BuyersPaid;

    // Modifiers

    modifier onlySeller(){
        require(seller == msg.sender , "Only the seller can manage funds");
        _;
    }

    modifier soldOut(){
        require(numTicketsSold < maxOccupancy , "All tickets have been sold");
        _;
    }

    // Events

    event Refund(address to , uint amount);

    // Functions

    constructor(uint _maxOccupancy , uint _price){
        seller = msg.sender;
        numTicketsSold = 0;
        maxOccupancy = _maxOccupancy;
        price = _price;
    }

    function buyTickets(string memory email , uint numTickets) public payable soldOut{
        BuyersPaid[msg.sender].email = email;
        
        uint sellTickets = numTickets;
        if(sellTickets + numTicketsSold > maxOccupancy){
            sellTickets = maxOccupancy - numTicketsSold;
        }

        uint payAmount = sellTickets * price;
        if(msg.value < payAmount){
            sellTickets = msg.value / price;
            payAmount = sellTickets * price;
        }

        BuyersPaid[msg.sender].totalPrice += msg.value;
        numTicketsSold += sellTickets;
        BuyersPaid[msg.sender].numTickets += sellTickets;
        
        if(payAmount < msg.value){
            refundTicket(payable(msg.sender));
            emit Refund(msg.sender , msg.value - payAmount);
        }
    }

    function withdrawFunds() public onlySeller{
        payable(msg.sender).transfer(address(this).balance);
    }

    function refundTicket(address buyer) public onlySeller{
        uint refund = BuyersPaid[buyer].totalPrice - price * BuyersPaid[buyer].numTickets;
        require(refund > 0 , "No funds to refund");
        payable(buyer).transfer(refund);
        BuyersPaid[buyer].totalPrice -= refund;
    }

    function getBuyerAmountPaid(address buyer) public view returns (uint){
        return BuyersPaid[buyer].totalPrice;
    }

    function kill() public onlySeller{
        selfdestruct(payable(msg.sender));
    }
}
